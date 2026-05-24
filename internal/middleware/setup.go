package middleware

import (
	"net/http"
	"strings"

	"github.com/jakovic/api-docs-portal/internal/models"
)

func RequireSetup(users *models.UserStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/static/") {
				next.ServeHTTP(w, r)
				return
			}

			if !users.HasUsers() && r.URL.Path != "/setup" {
				http.Redirect(w, r, "/setup", http.StatusSeeOther)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
