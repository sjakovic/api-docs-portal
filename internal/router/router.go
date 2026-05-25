package router

import (
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jakovic/api-docs-portal/internal/config"
	"github.com/jakovic/api-docs-portal/internal/handlers"
	"github.com/jakovic/api-docs-portal/internal/middleware"
	"github.com/jakovic/api-docs-portal/internal/models"
)

func New(cfg *config.Config, tmpl *template.Template, users *models.UserStore, docs *models.DocStore, permissions *models.PermissionStore, settings *models.SettingsStore, staticFS http.FileSystem) chi.Router {
	r := chi.NewRouter()

	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(middleware.RequireSetup(users))

	auth := middleware.NewAuthMiddleware(cfg.JWTSecret, users)

	setupHandler := handlers.NewSetupHandler(users, settings, auth, cfg.JWTExpiry, tmpl)
	authHandler := handlers.NewAuthHandler(users, auth, cfg.JWTExpiry, tmpl)
	adminUserHandler := handlers.NewAdminUserHandler(users, tmpl)
	adminDocHandler := handlers.NewAdminDocHandler(docs, tmpl)
	adminPermHandler := handlers.NewAdminPermissionHandler(permissions, users, docs, tmpl)
	adminSettingsHandler := handlers.NewAdminSettingsHandler(settings, tmpl)
	docHandler := handlers.NewDocHandler(docs, permissions, tmpl)
	contentHandler := handlers.NewContentHandler(docs, permissions)

	// Static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(staticFS)))

	// Setup (first run only)
	r.Get("/setup", setupHandler.SetupPage)
	r.Post("/setup", setupHandler.Setup)

	// Public routes
	r.Get("/login", authHandler.LoginPage)
	r.Post("/api/auth/login", authHandler.Login)
	r.Post("/api/auth/logout", authHandler.Logout)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireAuth)

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/docs", http.StatusSeeOther)
		})

		r.Get("/change-password", authHandler.ChangePasswordPage)
		r.Post("/change-password", authHandler.ChangePassword)

		r.Get("/docs", docHandler.Listing)
		r.Get("/docs/{slug}", docHandler.View)
		r.Get("/api/docs/{slug}/content", contentHandler.ServeContent)

		// Admin routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAdmin)

			r.Get("/admin", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/admin/docs", http.StatusSeeOther)
			})

			r.Get("/admin/users", adminUserHandler.List)
			r.Post("/admin/users", adminUserHandler.Create)
			r.Put("/admin/users/{id}", adminUserHandler.Update)
			r.Post("/admin/users/{id}", adminUserHandler.Update)
			r.Delete("/admin/users/{id}", adminUserHandler.Delete)
			r.Post("/admin/users/{id}/delete", adminUserHandler.Delete)

			r.Get("/admin/docs", adminDocHandler.List)
			r.Post("/admin/docs", adminDocHandler.Create)
			r.Get("/admin/docs/{id}/edit", adminDocHandler.Edit)
			r.Put("/admin/docs/{id}", adminDocHandler.Update)
			r.Post("/admin/docs/{id}", adminDocHandler.Update)
			r.Delete("/admin/docs/{id}", adminDocHandler.Delete)
			r.Post("/admin/docs/{id}/delete", adminDocHandler.Delete)

			r.Get("/admin/docs/{id}/permissions", adminPermHandler.DocPermissions)
			r.Post("/admin/docs/{id}/permissions", adminPermHandler.SetDocPermissions)

			r.Get("/admin/settings", adminSettingsHandler.Page)
			r.Post("/admin/settings", adminSettingsHandler.Save)
		})
	})

	return r
}
