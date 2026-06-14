package models

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type AuditLog struct {
	ID          uint      `gorm:"primaryKey"`
	ActorUserID uint      `gorm:"not null;index"`
	ActorRole   string    `gorm:"size:80;not null;index"`
	Action      string    `gorm:"size:120;not null;index"`
	EntityType  string    `gorm:"size:120;not null;index"`
	EntityID    string    `gorm:"size:80;not null;index"`
	BeforeJSON  string    `gorm:"type:jsonb"`
	AfterJSON   string    `gorm:"type:jsonb"`
	IPAddress   string    `gorm:"size:80"`
	UserAgent   string    `gorm:"size:255"`
	CreatedAt   time.Time `gorm:"not null;index"`

	Actor User `gorm:"foreignKey:ActorUserID;constraint:OnDelete:RESTRICT"`
}

func (a *AuditLog) BeforeSave(_ *gorm.DB) error {
	a.ActorRole = strings.TrimSpace(a.ActorRole)
	a.Action = strings.TrimSpace(a.Action)
	a.EntityType = strings.TrimSpace(a.EntityType)
	a.EntityID = strings.TrimSpace(a.EntityID)
	a.IPAddress = strings.TrimSpace(a.IPAddress)
	a.UserAgent = strings.TrimSpace(a.UserAgent)

	if a.ActorUserID == 0 || a.ActorRole == "" || a.Action == "" || a.EntityType == "" || a.EntityID == "" {
		return errors.New("audit log requires actor_user_id, actor_role, action, entity_type and entity_id")
	}

	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now().UTC()
	}

	return nil
}
