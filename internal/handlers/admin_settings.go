package handlers

import (
	"html/template"
	"log/slog"
	"net/http"

	"github.com/jakovic/api-docs-portal/internal/middleware"
	"github.com/jakovic/api-docs-portal/internal/models"
)

type AdminSettingsHandler struct {
	settings *models.SettingsStore
	tmpl     *template.Template
}

func NewAdminSettingsHandler(settings *models.SettingsStore, tmpl *template.Template) *AdminSettingsHandler {
	return &AdminSettingsHandler{settings: settings, tmpl: tmpl}
}

func (h *AdminSettingsHandler) Page(w http.ResponseWriter, r *http.Request) {
	h.tmpl.ExecuteTemplate(w, "admin_settings.html", map[string]interface{}{
		"User":      middleware.UserFromContext(r.Context()),
		"SiteTitle": h.settings.SiteTitle(),
		"Success":   r.URL.Query().Get("success"),
		"Error":     r.URL.Query().Get("error"),
	})
}

func (h *AdminSettingsHandler) Save(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/settings?error=Invalid+request", http.StatusSeeOther)
		return
	}

	siteTitle := r.FormValue("site_title")
	if siteTitle == "" {
		http.Redirect(w, r, "/admin/settings?error=Site+title+cannot+be+empty", http.StatusSeeOther)
		return
	}

	if err := h.settings.Set("site_title", siteTitle); err != nil {
		slog.Error("save site title", "error", err)
		http.Redirect(w, r, "/admin/settings?error=Failed+to+save+settings", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/settings?success=Settings+saved", http.StatusSeeOther)
}
