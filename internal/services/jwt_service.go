package services

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"kslasbackend/internal/config"
	"kslasbackend/internal/database/models"
)

type JWTService struct {
	secret    []byte
	issuer    string
	expiresIn time.Duration
}

type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	UUID     string `json:"uuid"`
	UserType string `json:"user_type"`
	jwt.RegisteredClaims
}

func NewJWTService(cfg config.Config) *JWTService {
	return &JWTService{
		secret:    []byte(cfg.JWTSecret),
		issuer:    cfg.AppName,
		expiresIn: time.Duration(cfg.JWTExpiresHours) * time.Hour,
	}
}

func (s *JWTService) GenerateToken(user *models.User) (string, error) {
	now := time.Now().UTC()

	claims := JWTClaims{
		UserID:   user.ID,
		UUID:     user.UUID,
		UserType: string(user.UserType),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			Subject:   fmt.Sprintf("%d", user.ID),
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiresIn)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *JWTService) ParseToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		if token.Method == nil || token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method")
		}

		return s.secret, nil
	}, jwt.WithIssuer(s.issuer))
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func (s *JWTService) ExpiresInSeconds() int {
	return int(s.expiresIn.Seconds())
}
