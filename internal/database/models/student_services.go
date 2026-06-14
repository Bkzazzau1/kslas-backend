package models

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

type SupportTicketStatus string

const (
	SupportTicketOpen      SupportTicketStatus = "open"
	SupportTicketAssigned  SupportTicketStatus = "assigned"
	SupportTicketEscalated SupportTicketStatus = "escalated"
	SupportTicketResolved  SupportTicketStatus = "resolved"
	SupportTicketClosed    SupportTicketStatus = "closed"
)

func (s SupportTicketStatus) Valid() bool {
	switch s {
	case SupportTicketOpen, SupportTicketAssigned, SupportTicketEscalated, SupportTicketResolved, SupportTicketClosed:
		return true
	default:
		return false
	}
}

type SupportTicket struct {
	ID             uint                `gorm:"primaryKey"`
	TicketNo       string              `gorm:"size:40;uniqueIndex;not null"`
	StudentID      uint                `gorm:"not null;index"`
	Category       string              `gorm:"size:80;not null;index"`
	Subject        string              `gorm:"size:180;not null"`
	Description    string              `gorm:"type:text;not null"`
	Priority       string              `gorm:"size:30;not null;default:normal;index"`
	Status         SupportTicketStatus `gorm:"size:30;not null;default:open;index"`
	CurrentOwnerID *uint               `gorm:"index"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ResolvedAt     *time.Time
	ClosedAt       *time.Time

	Student      User  `gorm:"foreignKey:StudentID;constraint:OnDelete:CASCADE"`
	CurrentOwner *User `gorm:"foreignKey:CurrentOwnerID;constraint:OnDelete:SET NULL"`
}

func (t *SupportTicket) BeforeSave(_ *gorm.DB) error {
	t.TicketNo = strings.ToUpper(strings.TrimSpace(t.TicketNo))
	t.Category = strings.TrimSpace(t.Category)
	t.Subject = strings.TrimSpace(t.Subject)
	t.Description = strings.TrimSpace(t.Description)
	t.Priority = strings.ToLower(strings.TrimSpace(t.Priority))

	if t.StudentID == 0 || t.TicketNo == "" || t.Category == "" || t.Subject == "" || t.Description == "" {
		return errors.New("support ticket requires ticket_no, student_id, category, subject and description")
	}

	if t.Priority == "" {
		t.Priority = "normal"
	}

	if t.Status == "" {
		t.Status = SupportTicketOpen
	}

	if !t.Status.Valid() {
		return errors.New("invalid support ticket status")
	}

	return nil
}

type SupportTicketMessage struct {
	ID         uint      `gorm:"primaryKey"`
	TicketID   uint      `gorm:"not null;index"`
	SenderID   uint      `gorm:"not null;index"`
	SenderType string    `gorm:"size:30;not null;index"`
	Message    string    `gorm:"type:text;not null"`
	CreatedAt  time.Time `gorm:"not null;index"`

	Ticket SupportTicket `gorm:"foreignKey:TicketID;constraint:OnDelete:CASCADE"`
	Sender User          `gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE"`
}

func (m *SupportTicketMessage) BeforeSave(_ *gorm.DB) error {
	m.SenderType = strings.ToLower(strings.TrimSpace(m.SenderType))
	m.Message = strings.TrimSpace(m.Message)

	if m.TicketID == 0 || m.SenderID == 0 || m.SenderType == "" || m.Message == "" {
		return errors.New("support ticket message requires ticket_id, sender_id, sender_type and message")
	}

	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now().UTC()
	}

	return nil
}

type SupportTicketFile struct {
	ID        uint      `gorm:"primaryKey"`
	TicketID  uint      `gorm:"not null;index"`
	FileID    uint      `gorm:"not null;index"`
	Purpose   string    `gorm:"size:80;not null"`
	CreatedAt time.Time `gorm:"not null;index"`

	Ticket SupportTicket `gorm:"foreignKey:TicketID;constraint:OnDelete:CASCADE"`
}

func (f *SupportTicketFile) BeforeSave(_ *gorm.DB) error {
	f.Purpose = strings.TrimSpace(f.Purpose)
	if f.TicketID == 0 || f.FileID == 0 || f.Purpose == "" {
		return errors.New("support ticket file requires ticket_id, file_id and purpose")
	}
	if f.CreatedAt.IsZero() {
		f.CreatedAt = time.Now().UTC()
	}
	return nil
}

type InternshipProfileStatus string

const (
	InternshipProfileDraft       InternshipProfileStatus = "draft"
	InternshipProfileUnderReview InternshipProfileStatus = "under_review"
	InternshipProfileApproved    InternshipProfileStatus = "approved"
	InternshipProfileRejected    InternshipProfileStatus = "rejected"
)

func (s InternshipProfileStatus) Valid() bool {
	switch s {
	case InternshipProfileDraft, InternshipProfileUnderReview, InternshipProfileApproved, InternshipProfileRejected:
		return true
	default:
		return false
	}
}

type InternshipProfile struct {
	ID                uint                    `gorm:"primaryKey"`
	StudentID         uint                    `gorm:"not null;uniqueIndex"`
	ProgrammeID       uint                    `gorm:"not null;index"`
	PreferredIndustry string                  `gorm:"size:120"`
	OrganizationName  string                  `gorm:"size:160"`
	OrganizationEmail string                  `gorm:"size:160"`
	OrganizationPhone string                  `gorm:"size:40"`
	OrganizationAddr  string                  `gorm:"type:text"`
	StartDate         *time.Time
	EndDate           *time.Time
	Status            InternshipProfileStatus `gorm:"size:40;not null;default:draft;index"`
	CreatedAt         time.Time
	UpdatedAt         time.Time

	Student   User      `gorm:"foreignKey:StudentID;constraint:OnDelete:CASCADE"`
	Programme Programme `gorm:"foreignKey:ProgrammeID;constraint:OnDelete:CASCADE"`
}

func (p *InternshipProfile) BeforeSave(_ *gorm.DB) error {
	p.PreferredIndustry = strings.TrimSpace(p.PreferredIndustry)
	p.OrganizationName = strings.TrimSpace(p.OrganizationName)
	p.OrganizationEmail = strings.ToLower(strings.TrimSpace(p.OrganizationEmail))
	p.OrganizationPhone = strings.TrimSpace(p.OrganizationPhone)
	p.OrganizationAddr = strings.TrimSpace(p.OrganizationAddr)

	if p.StudentID == 0 || p.ProgrammeID == 0 {
		return errors.New("internship profile requires student_id and programme_id")
	}

	if p.Status == "" {
		p.Status = InternshipProfileDraft
	}

	if !p.Status.Valid() {
		return errors.New("invalid internship profile status")
	}

	return nil
}

type InternshipLetterRequest struct {
	ID            uint      `gorm:"primaryKey"`
	RequestNo     string    `gorm:"size:40;uniqueIndex;not null"`
	StudentID     uint      `gorm:"not null;index"`
	ProfileID     uint      `gorm:"not null;index"`
	Status        string    `gorm:"size:40;not null;default:pending;index"`
	IssuedFileURL string    `gorm:"size:255"`
	RequestedAt   time.Time `gorm:"not null;index"`
	UpdatedAt     time.Time

	Student User              `gorm:"foreignKey:StudentID;constraint:OnDelete:CASCADE"`
	Profile InternshipProfile `gorm:"foreignKey:ProfileID;constraint:OnDelete:CASCADE"`
}

func (r *InternshipLetterRequest) BeforeSave(_ *gorm.DB) error {
	r.RequestNo = strings.ToUpper(strings.TrimSpace(r.RequestNo))
	r.Status = strings.ToLower(strings.TrimSpace(r.Status))
	r.IssuedFileURL = strings.TrimSpace(r.IssuedFileURL)

	if r.RequestNo == "" || r.StudentID == 0 || r.ProfileID == 0 {
		return errors.New("internship letter request requires request_no, student_id and profile_id")
	}

	if r.Status == "" {
		r.Status = "pending"
	}

	if r.RequestedAt.IsZero() {
		r.RequestedAt = time.Now().UTC()
	}

	return nil
}

type InternshipLogbookEntry struct {
	ID              uint      `gorm:"primaryKey"`
	StudentID       uint      `gorm:"not null;index"`
	ProfileID       uint      `gorm:"not null;index"`
	WeekNo          uint      `gorm:"not null;index"`
	ActivitySummary string    `gorm:"type:text;not null"`
	FileURL         string    `gorm:"size:255"`
	Status          string    `gorm:"size:40;not null;default:submitted;index"`
	SubmittedAt     time.Time `gorm:"not null;index"`
	ReviewedBy      *uint     `gorm:"index"`
	ReviewedAt      *time.Time

	Student  User              `gorm:"foreignKey:StudentID;constraint:OnDelete:CASCADE"`
	Profile  InternshipProfile `gorm:"foreignKey:ProfileID;constraint:OnDelete:CASCADE"`
	Reviewer *User             `gorm:"foreignKey:ReviewedBy;constraint:OnDelete:SET NULL"`
}

func (e *InternshipLogbookEntry) BeforeSave(_ *gorm.DB) error {
	e.ActivitySummary = strings.TrimSpace(e.ActivitySummary)
	e.FileURL = strings.TrimSpace(e.FileURL)
	e.Status = strings.ToLower(strings.TrimSpace(e.Status))

	if e.StudentID == 0 || e.ProfileID == 0 || e.WeekNo == 0 || e.ActivitySummary == "" {
		return errors.New("internship logbook entry requires student_id, profile_id, week_no and activity_summary")
	}

	if e.Status == "" {
		e.Status = "submitted"
	}

	if e.SubmittedAt.IsZero() {
		e.SubmittedAt = time.Now().UTC()
	}

	return nil
}

type TranscriptRequest struct {
	ID               uint       `gorm:"primaryKey"`
	RequestNo        string     `gorm:"size:40;uniqueIndex;not null"`
	StudentID        uint       `gorm:"not null;index"`
	DeliveryMethod   string     `gorm:"size:80;not null"`
	RecipientName    string     `gorm:"size:180;not null"`
	RecipientContact string     `gorm:"size:220;not null"`
	Purpose          string     `gorm:"size:80;not null"`
	Status           string     `gorm:"size:40;not null;default:pending;index"`
	RequestedAt      time.Time  `gorm:"not null;index"`
	CompletedAt      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time

	Student User `gorm:"foreignKey:StudentID;constraint:OnDelete:CASCADE"`
}

func (r *TranscriptRequest) BeforeSave(_ *gorm.DB) error {
	r.RequestNo = strings.ToUpper(strings.TrimSpace(r.RequestNo))
	r.DeliveryMethod = strings.TrimSpace(r.DeliveryMethod)
	r.RecipientName = strings.TrimSpace(r.RecipientName)
	r.RecipientContact = strings.TrimSpace(r.RecipientContact)
	r.Purpose = strings.TrimSpace(r.Purpose)
	r.Status = strings.ToLower(strings.TrimSpace(r.Status))

	if r.RequestNo == "" || r.StudentID == 0 || r.DeliveryMethod == "" || r.RecipientName == "" || r.RecipientContact == "" || r.Purpose == "" {
		return errors.New("transcript request requires request_no, student_id, delivery_method, recipient_name, recipient_contact and purpose")
	}

	if r.Status == "" {
		r.Status = "pending"
	}

	if r.RequestedAt.IsZero() {
		r.RequestedAt = time.Now().UTC()
	}

	return nil
}

type GraduationMap struct {
	ID                 uint      `gorm:"primaryKey"`
	StudentID          uint      `gorm:"not null;uniqueIndex"`
	ProgrammeID        uint      `gorm:"not null;index"`
	CreditsRequired    uint      `gorm:"not null"`
	CreditsEarned      uint      `gorm:"not null;default:0"`
	CreditsInProgress  uint      `gorm:"not null;default:0"`
	CreditsRemaining   uint      `gorm:"not null;default:0"`
	CoursesCompleted   uint      `gorm:"not null;default:0"`
	CoursesRemaining   uint      `gorm:"not null;default:0"`
	Carryovers         uint      `gorm:"not null;default:0"`
	ProgressPercent    uint      `gorm:"not null;default:0"`
	ExpectedGraduation string    `gorm:"size:80"`
	GeneratedAt        time.Time `gorm:"not null;index"`
	CreatedAt          time.Time
	UpdatedAt          time.Time

	Student   User      `gorm:"foreignKey:StudentID;constraint:OnDelete:CASCADE"`
	Programme Programme `gorm:"foreignKey:ProgrammeID;constraint:OnDelete:CASCADE"`
}

func (g *GraduationMap) BeforeSave(_ *gorm.DB) error {
	g.ExpectedGraduation = strings.TrimSpace(g.ExpectedGraduation)

	if g.StudentID == 0 || g.ProgrammeID == 0 || g.CreditsRequired == 0 {
		return errors.New("graduation map requires student_id, programme_id and credits_required")
	}

	if g.ProgressPercent > 100 {
		g.ProgressPercent = 100
	}

	if g.GeneratedAt.IsZero() {
		g.GeneratedAt = time.Now().UTC()
	}

	return nil
}

type GraduationMapItem struct {
	ID              uint      `gorm:"primaryKey"`
	GraduationMapID uint      `gorm:"not null;index"`
	CourseID        *uint     `gorm:"index"`
	Code            string    `gorm:"size:30;not null;index"`
	Title           string    `gorm:"size:180;not null"`
	RequirementType string    `gorm:"size:60;not null;index"`
	Credits         uint      `gorm:"not null"`
	Status          string    `gorm:"size:60;not null;index"`
	CreatedAt       time.Time
	UpdatedAt       time.Time

	GraduationMap GraduationMap `gorm:"foreignKey:GraduationMapID;constraint:OnDelete:CASCADE"`
	Course        *Course       `gorm:"foreignKey:CourseID;constraint:OnDelete:SET NULL"`
}

func (i *GraduationMapItem) BeforeSave(_ *gorm.DB) error {
	i.Code = strings.ToUpper(strings.TrimSpace(i.Code))
	i.Title = strings.TrimSpace(i.Title)
	i.RequirementType = strings.TrimSpace(i.RequirementType)
	i.Status = strings.TrimSpace(i.Status)

	if i.GraduationMapID == 0 || i.Code == "" || i.Title == "" || i.RequirementType == "" || i.Credits == 0 || i.Status == "" {
		return errors.New("graduation map item requires graduation_map_id, code, title, requirement_type, credits and status")
	}

	return nil
}
