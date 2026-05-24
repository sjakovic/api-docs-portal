package handlers

import (
	"bytes"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yuin/goldmark"
	"github.com/jakovic/api-docs-portal/internal/middleware"
	"github.com/jakovic/api-docs-portal/internal/models"
)

type DocHandler struct {
	docs        *models.DocStore
	permissions *models.PermissionStore
	tmpl        *template.Template
}

func NewDocHandler(docs *models.DocStore, permissions *models.PermissionStore, tmpl *template.Template) *DocHandler {
	return &DocHandler{docs: docs, permissions: permissions, tmpl: tmpl}
}

func (h *DocHandler) Listing(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())

	var docs []*models.Doc
	var err error

	if user.Role == "admin" {
		docs, err = h.docs.ListActive()
	} else {
		docs, err = h.docs.ListForUser(user.ID)
	}

	if err != nil {
		slog.Error("list docs", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	h.tmpl.ExecuteTemplate(w, "docs_listing.html", map[string]interface{}{
		"User": user,
		"Docs": docs,
	})
}

func (h *DocHandler) View(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	user := middleware.UserFromContext(r.Context())

	doc, err := h.docs.GetBySlug(slug)
	if err != nil {
		http.Error(w, "Doc not found", http.StatusNotFound)
		return
	}

	if user.Role != "admin" {
		hasAccess, err := h.permissions.HasAccess(user.ID, doc.ID)
		if err != nil || !hasAccess {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	}

	data := map[string]interface{}{
		"User": user,
		"Doc":  doc,
	}

	if doc.DocType == "markdown" && doc.Content.Valid {
		var buf bytes.Buffer
		if err := goldmark.Convert([]byte(doc.Content.String), &buf); err != nil {
			slog.Error("render markdown", "error", err)
		} else {
			data["RenderedContent"] = template.HTML(buf.String())
		}
	}

	templateName := "docs_" + doc.DocType + ".html"
	h.tmpl.ExecuteTemplate(w, templateName, data)
}
