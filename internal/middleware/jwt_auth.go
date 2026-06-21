package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type StaffJWTClaims struct {
	StaffID uuid.UUID `json:"staff_id"`
	Role    string    `json:"role"`
	Roles   []string  `json:"roles"`
	jwt.RegisteredClaims
}

func JWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "kslas-dev-secret-change-in-production"
	}
	return []byte(secret)
}

func NewStaffToken(staffID uuid.UUID, primaryRole string, roles []string) (string, error) {
	if primaryRole != "" {
		roles = append(roles, primaryRole)
	}
	claims := StaffJWTClaims{
		StaffID: staffID,
		Role:    primaryRole,
		Roles:   uniqueRoles(roles),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   staffID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(12 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(JWTSecret())
}

func StaffClaimsFromBearerToken(authHeader string) (StaffClaims, bool) {
	if authHeader == "" {
		return StaffClaims{}, false
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return StaffClaims{}, false
	}

	claims := &StaffJWTClaims{}
	token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (any, error) {
		return JWTSecret(), nil
	})
	if err != nil || !token.Valid || claims.StaffID == uuid.Nil {
		return StaffClaims{}, false
	}
	return StaffClaims{ID: claims.StaffID, Role: claims.Role, Roles: uniqueRoles(claims.Roles)}, true
}
