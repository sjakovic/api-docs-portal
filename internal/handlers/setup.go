package handlers

import (
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/jakovic/api-docs-portal/internal/middleware"
	"github.com/jakovic/api-docs-portal/internal/models"
)

type SetupHandler struct {
	users    *models.UserStore
	settings *models.SettingsStore
	auth     *middleware.AuthMiddleware
	expiry   time.Duration
	tmpl     *template.Template
}

func NewSetupHandler(users *models.UserStore, settings *models.SettingsStore, auth *middleware.AuthMiddleware, expiry time.Duration, tmpl *template.Template) *SetupHandler {
	return &SetupHandler{users: users, settings: settings, auth: auth, expiry: expiry, tmpl: tmpl}
}

func (h *SetupHandler) SetupPage(w http.ResponseWriter, r *http.Request) {
	if h.users.HasUsers() {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	h.tmpl.ExecuteTemplate(w, "setup.html", map[string]interface{}{
		"Error": r.URL.Query().Get("error"),
	})
}

func (h *SetupHandler) Setup(w http.ResponseWriter, r *http.Request) {
	if h.users.HasUsers() {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/setup?error=Invalid+request", http.StatusSeeOther)
		return
	}

	siteTitle := r.FormValue("site_title")
	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	if name == "" || email == "" || password == "" {
		http.Redirect(w, r, "/setup?error=All+fields+are+required", http.StatusSeeOther)
		return
	}

	if len(password) < 8 {
		http.Redirect(w, r, "/setup?error=Password+must+be+at+least+8+characters", http.StatusSeeOther)
		return
	}

	if password != confirmPassword {
		http.Redirect(w, r, "/setup?error=Passwords+do+not+match", http.StatusSeeOther)
		return
	}

	if siteTitle != "" {
		if err := h.settings.Set("site_title", siteTitle); err != nil {
			slog.Error("save site title", "error", err)
		}
	}

	user, err := h.users.CreateAdmin(email, password, name)
	if err != nil {
		slog.Error("create admin", "error", err)
		http.Redirect(w, r, "/setup?error=Failed+to+create+admin+account", http.StatusSeeOther)
		return
	}

	token, err := h.auth.GenerateToken(user.ID, h.expiry)
	if err != nil {
		slog.Error("generate token", "error", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
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

	http.Redirect(w, r, "/docs", http.StatusSeeOther)
}
