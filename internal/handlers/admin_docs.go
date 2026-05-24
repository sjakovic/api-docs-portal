package handlers

import (
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jakovic/api-docs-portal/internal/middleware"
	"github.com/jakovic/api-docs-portal/internal/models"
)

type AdminDocHandler struct {
	docs *models.DocStore
	tmpl *template.Template
}

func NewAdminDocHandler(docs *models.DocStore, tmpl *template.Template) *AdminDocHandler {
	return &AdminDocHandler{docs: docs, tmpl: tmpl}
}

func (h *AdminDocHandler) List(w http.ResponseWriter, r *http.Request) {
	docs, err := h.docs.List()
	if err != nil {
		slog.Error("list docs", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	h.tmpl.ExecuteTemplate(w, "admin_docs.html", map[string]interface{}{
		"User":    middleware.UserFromContext(r.Context()),
		"Docs":    docs,
		"Success": r.URL.Query().Get("success"),
		"Error":   r.URL.Query().Get("error"),
	})
}

func (h *AdminDocHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/docs?error=Invalid+request", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	slug := r.FormValue("slug")
	description := r.FormValue("description")
	docType := r.FormValue("doc_type")
	content := r.FormValue("content")
	externalURL := r.FormValue("external_url")
	version := r.FormValue("version")
	sortOrder, _ := strconv.Atoi(r.FormValue("sort_order"))

	if name == "" || docType == "" {
		http.Redirect(w, r, "/admin/docs?error=Name+and+type+required", http.StatusSeeOther)
		return
	}

	if _, err := h.docs.Create(name, slug, description, docType, content, externalURL, version, sortOrder); err != nil {
		slog.Error("create doc", "error", err)
		http.Redirect(w, r, "/admin/docs?error=Failed+to+create+doc", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/docs?success=Doc+created", http.StatusSeeOther)
}

func (h *AdminDocHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/docs?error=Invalid+request", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	slug := r.FormValue("slug")
	description := r.FormValue("description")
	docType := r.FormValue("doc_type")
	content := r.FormValue("content")
	externalURL := r.FormValue("external_url")
	version := r.FormValue("version")
	sortOrder, _ := strconv.Atoi(r.FormValue("sort_order"))
	isActive := r.FormValue("is_active") == "on" || r.FormValue("is_active") == "true"

	if err := h.docs.Update(id, name, slug, description, docType, content, externalURL, version, sortOrder, isActive); err != nil {
		slog.Error("update doc", "error", err)
		http.Redirect(w, r, "/admin/docs?error=Failed+to+update+doc", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/docs?success=Doc+updated", http.StatusSeeOther)
}

func (h *AdminDocHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.docs.Deactivate(id); err != nil {
		slog.Error("deactivate doc", "error", err)
		http.Redirect(w, r, "/admin/docs?error=Failed+to+deactivate+doc", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/docs?success=Doc+deactivated", http.StatusSeeOther)
}
