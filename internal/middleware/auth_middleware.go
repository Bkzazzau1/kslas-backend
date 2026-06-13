package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"kslasbackend/internal/services"
)

type contextKey string

const jwtClaimsContextKey contextKey = "jwt_claims"

func AuthMiddleware(jwtService *services.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			if authHeader == "" {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "missing authorization header"})
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "invalid authorization format"})
				return
			}

			claims, err := jwtService.ParseToken(parts[1])
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "invalid or expired token"})
				return
			}

			ctx := context.WithValue(r.Context(), jwtClaimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserIDFromContext(ctx context.Context) (uint, bool) {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		return 0, false
	}

	return claims.UserID, true
}

func ClaimsFromContext(ctx context.Context) (*services.JWTClaims, bool) {
	claims, ok := ctx.Value(jwtClaimsContextKey).(*services.JWTClaims)
	return claims, ok
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
