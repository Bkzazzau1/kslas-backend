package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type contextKey string

const staffContextKey contextKey = "staff_claims"

type StaffClaims struct {
	ID    uuid.UUID `json:"id"`
	Role  string    `json:"role"`
	Roles []string  `json:"roles"`
}

func RequireRoles(allowedRoles ...string) func(http.Handler) http.Handler {
	allowed := map[string]bool{}
	for _, role := range allowedRoles {
		allowed[strings.TrimSpace(role)] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := StaffClaimsFromHeaders(r)
			if !ok {
				writeMiddlewareError(w, http.StatusUnauthorized, "staff authentication headers are required")
				return
			}
			if len(allowed) > 0 && !claims.HasAnyRole(allowed) {
				writeMiddlewareError(w, http.StatusForbidden, "staff role is not allowed for this action")
				return
			}
			ctx := context.WithValue(r.Context(), staffContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func StaffClaimsFromHeaders(r *http.Request) (StaffClaims, bool) {
	idText := strings.TrimSpace(r.Header.Get("X-Staff-ID"))
	if idText == "" {
		return StaffClaims{}, false
	}
	id, err := uuid.Parse(idText)
	if err != nil {
		return StaffClaims{}, false
	}

	role := strings.TrimSpace(r.Header.Get("X-Staff-Role"))
	roles := parseRoles(r.Header.Get("X-Staff-Roles"))
	if role != "" {
		roles = append(roles, role)
	}
	if len(roles) == 0 {
		return StaffClaims{}, false
	}

	return StaffClaims{ID: id, Role: role, Roles: uniqueRoles(roles)}, true
}

func StaffClaimsFromContext(ctx context.Context) (StaffClaims, bool) {
	claims, ok := ctx.Value(staffContextKey).(StaffClaims)
	return claims, ok
}

func (c StaffClaims) HasAnyRole(allowed map[string]bool) bool {
	for _, role := range c.Roles {
		if allowed[role] {
			return true
		}
	}
	return false
}

func parseRoles(value string) []string {
	parts := strings.Split(value, ",")
	roles := []string{}
	for _, part := range parts {
		role := strings.TrimSpace(part)
		if role != "" {
			roles = append(roles, role)
		}
	}
	return roles
}

func uniqueRoles(roles []string) []string {
	seen := map[string]bool{}
	unique := []string{}
	for _, role := range roles {
		if !seen[role] {
			seen[role] = true
			unique = append(unique, role)
		}
	}
	return unique
}

func writeMiddlewareError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": message})
}
