package middleware

import (
	"encoding/json"
	"net/http"

	"kslasbackend/internal/rbac"
	"kslasbackend/internal/services"
)

type ScopeResolver func(*http.Request) (*rbac.Scope, error)

func RequirePermission(permissionService *services.PermissionService, permissionCode string, scopeResolver ScopeResolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				writePermissionJSON(w, http.StatusUnauthorized, "unauthenticated user")
				return
			}

			var target *rbac.Scope
			if scopeResolver != nil {
				scope, err := scopeResolver(r)
				if err != nil {
					writePermissionJSON(w, http.StatusBadRequest, err.Error())
					return
				}

				target = scope
			}

			allowed, err := permissionService.UserHasPermission(r.Context(), userID, permissionCode, target)
			if err != nil {
				writePermissionJSON(w, http.StatusInternalServerError, "failed to check permission")
				return
			}

			if !allowed {
				writePermissionJSON(w, http.StatusForbidden, "permission denied")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func writePermissionJSON(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": message,
	})
}
