package services

import (
	"context"
	"strings"
)

func (s *AuthService) ChangeCredential(ctx context.Context, userID uint, currentValue, newValue string) error {
	currentValue = strings.TrimSpace(currentValue)
	newValue = strings.TrimSpace(newValue)
	if currentValue == "" || newValue == "" {
		return ValidationError{Message: "current and new values are required"}
	}
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if err := s.passwordService.Verify(currentValue, user.PasswordHash); err != nil {
		return ErrInvalidCredentials
	}
	hash, err := s.passwordService.Hash(newValue)
	if err != nil {
		return ValidationError{Message: err.Error()}
	}
	return s.repo.SaveUserCredentialHash(ctx, userID, hash)
}
