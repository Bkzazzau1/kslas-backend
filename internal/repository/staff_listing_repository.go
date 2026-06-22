package repository

import (
	"context"

	"gorm.io/gorm"

	"kslasbackend/internal/database/models"
)

func (r *TeachingRepository) ListStaff(ctx context.Context) ([]models.User, error) {
	var users []models.User
	err := r.db.WithContext(ctx).
		Preload("UserRoles", func(db *gorm.DB) *gorm.DB {
			return db.Order("is_primary DESC, created_at ASC")
		}).
		Preload("UserRoles.Role").
		Where("user_type = ?", models.UserTypeStaff).
		Order("created_at DESC").
		Find(&users).Error
	return users, err
}

func (r *TeachingRepository) GetStaffUser(ctx context.Context, staffUserID uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Preload("UserRoles", func(db *gorm.DB) *gorm.DB {
			return db.Order("is_primary DESC, created_at ASC")
		}).
		Preload("UserRoles.Role").
		Where("id = ? AND user_type = ?", staffUserID, models.UserTypeStaff).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *TeachingRepository) UpdateStaffStatus(ctx context.Context, staffUserID uint, status models.UserStatus) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ? AND user_type = ?", staffUserID, models.UserTypeStaff).
		UpdateColumn("status", status).
		Error
}
