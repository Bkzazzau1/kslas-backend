package repository

import (
	"context"

	"kslasbackend/internal/database/models"
)

func (r *TeachingRepository) SaveStaffCredentialHash(ctx context.Context, staffUserID uint, credentialHash string) error {
	updates := map[string]any{"password_hash": credentialHash}
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ? AND user_type = ?", staffUserID, models.UserTypeStaff).
		Updates(updates).
		Error
}

func (r *AuthRepository) SaveUserCredentialHash(ctx context.Context, userID uint, credentialHash string) error {
	updates := map[string]any{"password_hash": credentialHash}
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Updates(updates).
		Error
}
