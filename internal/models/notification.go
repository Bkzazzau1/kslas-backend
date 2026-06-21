package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Notification struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	RecipientID uuid.UUID     `json:"recipient_id" gorm:"type:uuid;not null;index"`
	SenderID    *uuid.UUID    `json:"sender_id" gorm:"type:uuid;index"`
	Title       string        `json:"title" gorm:"size:200;not null"`
	Message     string        `json:"message"`
	Category    string        `json:"category" gorm:"size:60;default:general;index"`
	Priority    string        `json:"priority" gorm:"size:20;default:normal;index"`
	ActionURL   string        `json:"action_url" gorm:"size:255"`
	ResourceType string       `json:"resource_type" gorm:"size:80;index"`
	ResourceID   *uuid.UUID   `json:"resource_id" gorm:"type:uuid;index"`
	Metadata    datatypes.JSON `json:"metadata"`
	ReadAt      *time.Time    `json:"read_at"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	if n.Category == "" {
		n.Category = "general"
	}
	if n.Priority == "" {
		n.Priority = "normal"
	}
	return nil
}
