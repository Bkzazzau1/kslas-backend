package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"gorm.io/gorm"

	"kslasbackend/internal/config"
	"kslasbackend/internal/database/models"
	"kslasbackend/internal/services"
)

func SeedBootstrapAdmin(ctx context.Context, db *gorm.DB, cfg config.Config, passwordService *services.PasswordService) error {
	if cfg.BootstrapAdmin.Email == "" || cfg.BootstrapAdmin.Password == "" {
		return nil
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var admin models.User
		err := tx.Where("email = ?", strings.ToLower(cfg.BootstrapAdmin.Email)).First(&admin).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			hash, hashErr := passwordService.Hash(cfg.BootstrapAdmin.Password)
			if hashErr != nil {
				return fmt.Errorf("hash bootstrap admin password: %w", hashErr)
			}

			admin = models.User{
				FirstName:    cfg.BootstrapAdmin.FirstName,
				LastName:     cfg.BootstrapAdmin.LastName,
				Email:        strings.ToLower(cfg.BootstrapAdmin.Email),
				PasswordHash: hash,
				Status:       models.UserStatusActive,
				UserType:     models.UserTypeStaff,
			}

			if createErr := tx.Create(&admin).Error; createErr != nil {
				return fmt.Errorf("create bootstrap admin: %w", createErr)
			}

			log.Printf("bootstrap admin created for %s", admin.Email)
		case err != nil:
			return fmt.Errorf("lookup bootstrap admin: %w", err)
		}

		var role models.Role
		if err := tx.Where("code = ?", "system_admin").First(&role).Error; err != nil {
			return fmt.Errorf("load system_admin role: %w", err)
		}

		userRole := models.UserRole{
			UserID:    admin.ID,
			RoleID:    role.ID,
			ScopeType: models.ScopeSchool,
			IsPrimary: true,
		}

		if err := tx.Where("user_id = ? AND role_id = ? AND scope_key = ?", admin.ID, role.ID, "school:*").
			FirstOrCreate(&userRole).Error; err != nil {
			return fmt.Errorf("ensure bootstrap admin role: %w", err)
		}

		return nil
	})
}
