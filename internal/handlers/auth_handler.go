package handlers

import (
	"errors"
	"log"
	"net/http"

	"kslasbackend/internal/dto"
	"kslasbackend/internal/middleware"
	"kslasbackend/internal/services"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	response, err := h.authService.Login(r.Context(), req.Identity, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrInvalidCredentials), errors.Is(err, services.ErrInactiveAccount):
			writeError(w, http.StatusUnauthorized, "invalid credentials")
		default:
			log.Printf("login failed internal error: %v", err)
			writeError(w, http.StatusInternalServerError, "login failed")
		}
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return
	}

	user, err := h.authService.CurrentUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch current user")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user": user,
	})
}
