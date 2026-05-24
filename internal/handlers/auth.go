package handlers

import (
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/jakovic/api-docs-portal/internal/middleware"
	"github.com/jakovic/api-docs-portal/internal/models"
)

type AuthHandler struct {
	users    *models.UserStore
	auth     *middleware.AuthMiddleware
	expiry   time.Duration
	tmpl     *template.Template
}

func NewAuthHandler(users *models.UserStore, auth *middleware.AuthMiddleware, expiry time.Duration, tmpl *template.Template) *AuthHandler {
	return &AuthHandler{users: users, auth: auth, expiry: expiry, tmpl: tmpl}
}

func (h *AuthHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	h.tmpl.ExecuteTemplate(w, "login.html", map[string]interface{}{
		"Error": r.URL.Query().Get("error"),
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/login?error=Invalid+request", http.StatusSeeOther)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := h.users.GetByEmail(email)
	if err != nil || !h.users.CheckPassword(user, password) {
		http.Redirect(w, r, "/login?error=Invalid+email+or+password", http.StatusSeeOther)
		return
	}

	if !user.IsActive {
		http.Redirect(w, r, "/login?error=Account+disabled", http.StatusSeeOther)
		return
	}

	token, err := h.auth.GenerateToken(user.ID, h.expiry)
	if err != nil {
		slog.Error("generate token", "error", err)
		http.Redirect(w, r, "/login?error=Internal+error", http.StatusSeeOther)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(h.expiry.Seconds()),
	})

	if user.MustChangePassword {
		http.Redirect(w, r, "/change-password", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/docs", http.StatusSeeOther)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *AuthHandler) ChangePasswordPage(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	h.tmpl.ExecuteTemplate(w, "change_password.html", map[string]interface{}{
		"User":  user,
		"Error": r.URL.Query().Get("error"),
	})
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/change-password?error=Invalid+request", http.StatusSeeOther)
		return
	}

	newPassword := r.FormValue("new_password")
	confirmPassword := r.FormValue("confirm_password")

	if len(newPassword) < 8 {
		http.Redirect(w, r, "/change-password?error=Password+must+be+at+least+8+characters", http.StatusSeeOther)
		return
	}

	if newPassword != confirmPassword {
		http.Redirect(w, r, "/change-password?error=Passwords+do+not+match", http.StatusSeeOther)
		return
	}

	if err := h.users.UpdatePassword(user.ID, newPassword); err != nil {
		slog.Error("update password", "error", err)
		http.Redirect(w, r, "/change-password?error=Failed+to+update+password", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/docs", http.StatusSeeOther)
}
