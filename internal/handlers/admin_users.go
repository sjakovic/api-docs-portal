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

type AdminUserHandler struct {
	users *models.UserStore
	tmpl  *template.Template
}

func NewAdminUserHandler(users *models.UserStore, tmpl *template.Template) *AdminUserHandler {
	return &AdminUserHandler{users: users, tmpl: tmpl}
}

func (h *AdminUserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.users.List()
	if err != nil {
		slog.Error("list users", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	h.tmpl.ExecuteTemplate(w, "admin_users.html", map[string]interface{}{
		"User":    middleware.UserFromContext(r.Context()),
		"Users":   users,
		"Success": r.URL.Query().Get("success"),
		"Error":   r.URL.Query().Get("error"),
	})
}

func (h *AdminUserHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/users?error=Invalid+request", http.StatusSeeOther)
		return
	}

	email := r.FormValue("email")
	name := r.FormValue("name")
	password := r.FormValue("password")
	role := r.FormValue("role")

	if email == "" || name == "" || password == "" {
		http.Redirect(w, r, "/admin/users?error=All+fields+required", http.StatusSeeOther)
		return
	}
	if role != "admin" && role != "viewer" {
		role = "viewer"
	}

	if _, err := h.users.Create(email, password, name, role); err != nil {
		slog.Error("create user", "error", err)
		http.Redirect(w, r, "/admin/users?error=Failed+to+create+user", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/users?success=User+created", http.StatusSeeOther)
}

func (h *AdminUserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/users?error=Invalid+request", http.StatusSeeOther)
		return
	}

	email := r.FormValue("email")
	name := r.FormValue("name")
	role := r.FormValue("role")
	isActive := r.FormValue("is_active") == "on" || r.FormValue("is_active") == "true"

	if err := h.users.Update(id, email, name, role, isActive); err != nil {
		slog.Error("update user", "error", err)
		http.Redirect(w, r, "/admin/users?error=Failed+to+update+user", http.StatusSeeOther)
		return
	}

	newPassword := r.FormValue("password")
	if newPassword != "" {
		if err := h.users.UpdatePassword(id, newPassword); err != nil {
			slog.Error("update user password", "error", err)
		}
	}

	http.Redirect(w, r, "/admin/users?success=User+updated", http.StatusSeeOther)
}

func (h *AdminUserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	currentUser := middleware.UserFromContext(r.Context())
	if currentUser.ID == id {
		http.Redirect(w, r, "/admin/users?error=Cannot+deactivate+yourself", http.StatusSeeOther)
		return
	}

	if err := h.users.Deactivate(id); err != nil {
		slog.Error("deactivate user", "error", err)
		http.Redirect(w, r, "/admin/users?error=Failed+to+deactivate+user", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/users?success=User+deactivated", http.StatusSeeOther)
}
