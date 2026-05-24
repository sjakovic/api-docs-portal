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

type AdminPermissionHandler struct {
	permissions *models.PermissionStore
	users       *models.UserStore
	docs        *models.DocStore
	tmpl        *template.Template
}

func NewAdminPermissionHandler(permissions *models.PermissionStore, users *models.UserStore, docs *models.DocStore, tmpl *template.Template) *AdminPermissionHandler {
	return &AdminPermissionHandler{permissions: permissions, users: users, docs: docs, tmpl: tmpl}
}

func (h *AdminPermissionHandler) DocPermissions(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	doc, err := h.docs.GetByID(id)
	if err != nil {
		http.Error(w, "Doc not found", http.StatusNotFound)
		return
	}

	allUsers, err := h.users.List()
	if err != nil {
		slog.Error("list users", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	grantedUsers, err := h.permissions.GetDocUsers(doc.ID)
	if err != nil {
		slog.Error("get doc users", "error", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	grantedMap := make(map[int64]bool)
	for _, u := range grantedUsers {
		grantedMap[u.ID] = true
	}

	h.tmpl.ExecuteTemplate(w, "admin_permissions.html", map[string]interface{}{
		"User":       middleware.UserFromContext(r.Context()),
		"Doc":        doc,
		"AllUsers":   allUsers,
		"GrantedMap": grantedMap,
		"Success":    r.URL.Query().Get("success"),
		"Error":      r.URL.Query().Get("error"),
	})
}

func (h *AdminPermissionHandler) SetDocPermissions(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/docs/"+strconv.FormatInt(id, 10)+"/permissions?error=Invalid+request", http.StatusSeeOther)
		return
	}

	var userIDs []int64
	for _, v := range r.Form["user_ids"] {
		uid, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			continue
		}
		userIDs = append(userIDs, uid)
	}

	currentUser := middleware.UserFromContext(r.Context())
	if err := h.permissions.SetDocPermissions(id, userIDs, currentUser.ID); err != nil {
		slog.Error("set permissions", "error", err)
		http.Redirect(w, r, "/admin/docs/"+strconv.FormatInt(id, 10)+"/permissions?error=Failed+to+update+permissions", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/docs/"+strconv.FormatInt(id, 10)+"/permissions?success=Permissions+updated", http.StatusSeeOther)
}
