package repository

import (
"context"

"kslasbackend/internal/database/models"
)

func (r *TeachingRepository) SaveStaffCredentialHash(ctx context.Context, staffUserID uint, credentialHash string) error {
return r.db.WithContext(ctx).
Model(&models.User{}).
Where("id = ? AND user_type = ?", staffUserID, models.UserTypeStaff).
UpdateColumn("password_hash", credentialHash).
Error
}

func (r *AuthRepository) SaveUserCredentialHash(ctx context.Context, userID uint, credentialHash string) error {
return r.db.WithContext(ctx).
Model(&models.User{}).
Where("id = ?", userID).
UpdateColumn("password_hash", credentialHash).
Error
}
