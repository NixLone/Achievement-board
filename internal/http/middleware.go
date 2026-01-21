package http

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"firegoals/internal/auth"

	"github.com/golang-jwt/jwt/v5"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func (a *API) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && a.isOriginAllowed(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *API) isOriginAllowed(origin string) bool {
	if origin == "http://localhost:5173" {
		return true
	}
	if len(a.Origins) == 0 {
		return false
	}
	for _, allowed := range a.Origins {
		if allowed == "*" {
			return true
		}
		if strings.EqualFold(strings.TrimSpace(allowed), origin) {
			return true
		}
	}
	return false
}

func (a *API) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := auth.TokenFromRequest(r)
		if !ok {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing token")
			return
		}
		claims, err := a.Auth.ParseToken(token)
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				writeError(w, http.StatusUnauthorized, "TOKEN_EXPIRED", "Token expired")
				return
			}
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token")
			return
		}
		ctx := auth.WithUserID(r.Context(), claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
