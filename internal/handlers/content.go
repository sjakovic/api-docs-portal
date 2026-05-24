package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jakovic/api-docs-portal/internal/middleware"
	"github.com/jakovic/api-docs-portal/internal/models"
)

type ContentHandler struct {
	docs        *models.DocStore
	permissions *models.PermissionStore
}

func NewContentHandler(docs *models.DocStore, permissions *models.PermissionStore) *ContentHandler {
	return &ContentHandler{docs: docs, permissions: permissions}
}

func (h *ContentHandler) ServeContent(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	user := middleware.UserFromContext(r.Context())

	doc, err := h.docs.GetBySlug(slug)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if user.Role != "admin" {
		hasAccess, err := h.permissions.HasAccess(user.ID, doc.ID)
		if err != nil || !hasAccess {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	}

	switch doc.DocType {
	case "openapi":
		w.Header().Set("Content-Type", "application/json")
	case "markdown":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	default:
		w.Header().Set("Content-Type", "text/plain")
	}

	if doc.Content.Valid {
		w.Write([]byte(doc.Content.String))
	}
}
