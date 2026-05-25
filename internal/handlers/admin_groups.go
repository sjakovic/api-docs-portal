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

type AdminGroupHandler struct {
	groups *models.DocGroupStore
	tmpl   *template.Template
}

func NewAdminGroupHandler(groups *models.DocGroupStore, tmpl *template.Template) *AdminGroupHandler {
	return &AdminGroupHandler{groups: groups, tmpl: tmpl}
}

func (h *AdminGroupHandler) List(w http.ResponseWriter, r *http.Request) {
	groups, err := h.groups.List()
	if err != nil {
		slog.Error("list groups", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	h.tmpl.ExecuteTemplate(w, "admin_groups.html", map[string]interface{}{
		"User":    middleware.UserFromContext(r.Context()),
		"Groups":  groups,
		"Success": r.URL.Query().Get("success"),
		"Error":   r.URL.Query().Get("error"),
	})
}

func (h *AdminGroupHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/groups?error=Invalid+request", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	sortOrder, _ := strconv.Atoi(r.FormValue("sort_order"))

	if name == "" {
		http.Redirect(w, r, "/admin/groups?error=Name+is+required", http.StatusSeeOther)
		return
	}

	if _, err := h.groups.Create(name, sortOrder); err != nil {
		slog.Error("create group", "error", err)
		http.Redirect(w, r, "/admin/groups?error=Failed+to+create+group", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/groups?success=Group+created", http.StatusSeeOther)
}

func (h *AdminGroupHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/groups?error=Invalid+request", http.StatusSeeOther)
		return
	}

	name := r.FormValue("name")
	sortOrder, _ := strconv.Atoi(r.FormValue("sort_order"))

	if err := h.groups.Update(id, name, sortOrder); err != nil {
		slog.Error("update group", "error", err)
		http.Redirect(w, r, "/admin/groups?error=Failed+to+update+group", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/groups?success=Group+updated", http.StatusSeeOther)
}

func (h *AdminGroupHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := h.groups.Delete(id); err != nil {
		slog.Error("delete group", "error", err)
		http.Redirect(w, r, "/admin/groups?error=Failed+to+delete+group", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/groups?success=Group+deleted", http.StatusSeeOther)
}
