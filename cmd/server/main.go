package main

import (
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"github.com/jakovic/api-docs-portal/internal/config"
	"github.com/jakovic/api-docs-portal/internal/database"
	"github.com/jakovic/api-docs-portal/internal/models"
	"github.com/jakovic/api-docs-portal/internal/router"
	"github.com/jakovic/api-docs-portal/web"
)

var version = "dev"

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))
	slog.Info("api-docs-portal", "version", version)

	cfg := config.Load()

	db, err := database.Open(cfg.DBPath)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	users := models.NewUserStore(db)
	docs := models.NewDocStore(db)
	permissions := models.NewPermissionStore(db)
	settings := models.NewSettingsStore(db)

	funcMap := template.FuncMap{
		"siteTitle": settings.SiteTitle,
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(web.TemplateFS,
		"templates/layouts/*.html",
		"templates/auth/*.html",
		"templates/admin/*.html",
		"templates/docs/*.html",
	)
	if err != nil {
		slog.Error("failed to parse templates", "error", err)
		os.Exit(1)
	}

	staticFS, err := fs.Sub(web.StaticFS, "static")
	if err != nil {
		slog.Error("failed to create static fs", "error", err)
		os.Exit(1)
	}

	r := router.New(cfg, tmpl, users, docs, permissions, settings, http.FS(staticFS))

	slog.Info("server starting", "addr", cfg.Addr())
	if err := http.ListenAndServe(cfg.Addr(), r); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
