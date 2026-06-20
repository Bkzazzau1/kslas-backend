package repository

import (
	"context"
	"strings"
	"time"

	"gorm.io/gorm"

	"kslasbackend/internal/database/models"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) FindUserByIdentity(ctx context.Context, identity string) (*models.User, error) {
	identity = strings.TrimSpace(identity)

	var user models.User
	err := r.db.WithContext(ctx).
		Preload("UserRoles", func(db *gorm.DB) *gorm.DB {
			return db.Order("is_primary DESC, created_at ASC")
		}).
		Preload("UserRoles.Role").
		Where(
			"LOWER(email) = ? OR phone = ? OR matric_no = ? OR staff_id = ?",
			strings.ToLower(identity),
			identity,
			identity,
			identity,
		).
		First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *AuthRepository) FindUserByID(ctx context.Context, userID uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Preload("UserRoles", func(db *gorm.DB) *gorm.DB {
			return db.Order("is_primary DESC, created_at ASC")
		}).
		Preload("UserRoles.Role").
		First(&user, userID).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *AuthRepository) UpdateLastLogin(ctx context.Context, userID uint, at time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		UpdateColumn("last_login_at", at.UTC()).
		Error
}
