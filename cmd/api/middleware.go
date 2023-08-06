package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/kickbu2towski/brb-api/internal/data"
)

func (app *application) logRequest(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		app.logger.Printf("method: %s, path: %s, origin: %s", r.Method, r.URL.Path, r.Header.Get("Origin"))
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func (app *application) isAuthenticated(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("sessionID")
		if err != nil {
			switch {
			case errors.Is(err, http.ErrNoCookie):
				app.badRequestResponse(w, r, "unauthorized")
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		user, err := app.models.Users.GetUserForToken(context.Background(), cookie.Value, data.ScopeAuthentication)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		r = app.setUserContext(r, user)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// TODO: vary header
func (app *application) enableCORS(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("origin")
		allowedOrigins := app.config.cors.allowedOrigins

		if origin != "" {
			for _, v := range allowedOrigins {
				if v == origin {
					w.Header().Set("Access-Control-Allow-Origin", v)
					w.Header().Set("Access-Control-Allow-Credentials", "true")

					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
						w.Header().Set("Access-Control-Allow-Methods", "POST, PUT, DELETE")
						w.WriteHeader(http.StatusOK)
						return
					}

					break
				}
			}
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func enableCORS(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		origin := w.Header().Get("Origin")
		origins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), " ")

		for _, og := range origins {
			if og == origin {
				w.Header().Set("Access-Control-Allow-Origin", origin)

				// handling preflight request
				if r.Method == http.MethodOptions && w.Header().Get("Access-Control-Request-Method") != "" {
					w.Header().Set("Access-Control-Allow-Method", "PUT, PATCH, DELETE, POST")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
					w.WriteHeader(http.StatusOK)
				}

				break
			}
		}
	}
	return http.HandlerFunc(fn)
}
