package repository

import (
	"context"

	"gorm.io/gorm"

	"kslasbackend/internal/database/models"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(ctx context.Context, log *models.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *AuditRepository) ListByEntity(ctx context.Context, entityType string, entityID string, limit int) ([]models.AuditLog, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	var logs []models.AuditLog
	err := r.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

func (r *AuditRepository) ListByActor(ctx context.Context, actorUserID uint, limit int) ([]models.AuditLog, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	var logs []models.AuditLog
	err := r.db.WithContext(ctx).
		Where("actor_user_id = ?", actorUserID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}
