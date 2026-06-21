package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Department struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Name      string    `json:"name" gorm:"size:180;not null"`
	Code      string    `json:"code" gorm:"size:30;uniqueIndex;not null"`
	Faculty   string    `json:"faculty" gorm:"size:180"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (d *Department) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

type Course struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	Code        string     `json:"code" gorm:"size:40;uniqueIndex;not null"`
	Title       string     `json:"title" gorm:"size:220;not null"`
	DepartmentID uuid.UUID `json:"department_id" gorm:"type:uuid;not null"`
	Department  Department `json:"department" gorm:"foreignKey:DepartmentID"`
	Level       int        `json:"level" gorm:"default:100"`
	Semester    string     `json:"semester" gorm:"size:30;default:first"`
	CreditUnits int        `json:"credit_units" gorm:"default:3"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (c *Course) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
