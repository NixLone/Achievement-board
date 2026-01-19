package http

import (
	"log"
	"net/http"
	"time"

	"firegoals/internal/auth"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *API) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := auth.TokenFromRequest(r)
		if !ok {
			writeError(w, http.StatusUnauthorized, "missing token")
			return
		}
		claims, err := a.Auth.ParseToken(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		ctx := auth.WithUserID(r.Context(), claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
