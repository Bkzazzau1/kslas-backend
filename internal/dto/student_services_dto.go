package dto

import "time"

type GraduationMapResponse struct {
	StudentID          uint                         `json:"student_id"`
	ProgrammeID        uint                         `json:"programme_id"`
	CreditsRequired    uint                         `json:"credits_required"`
	CreditsEarned      uint                         `json:"credits_earned"`
	CreditsInProgress  uint                         `json:"credits_in_progress"`
	CreditsRemaining   uint                         `json:"credits_remaining"`
	CoursesCompleted   uint                         `json:"courses_completed"`
	CoursesRemaining   uint                         `json:"courses_remaining"`
	Carryovers         uint                         `json:"carryovers"`
	ProgressPercent    uint                         `json:"progress_percent"`
	ExpectedGraduation string                       `json:"expected_graduation"`
	GeneratedAt        time.Time                    `json:"generated_at"`
	Items              []GraduationMapItemResponse  `json:"items"`
}

type GraduationMapItemResponse struct {
	CourseID        *uint  `json:"course_id,omitempty"`
	Code            string `json:"code"`
	Title           string `json:"title"`
	RequirementType string `json:"requirement_type"`
	Credits         uint   `json:"credits"`
	Status          string `json:"status"`
}

type SupportTicketResponse struct {
	ID             uint       `json:"id"`
	TicketNo       string     `json:"ticket_no"`
	Category       string     `json:"category"`
	Subject        string     `json:"subject"`
	Description    string     `json:"description"`
	Priority       string     `json:"priority"`
	Status         string     `json:"status"`
	CurrentOwnerID *uint      `json:"current_owner_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
	ClosedAt       *time.Time `json:"closed_at,omitempty"`
}

type CreateSupportTicketRequest struct {
	Category    string `json:"category"`
	Subject     string `json:"subject"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
}

type AddSupportReplyRequest struct {
	Message string `json:"message"`
}

type TranscriptRequestResponse struct {
	ID               uint       `json:"id"`
	RequestNo        string     `json:"request_no"`
	DeliveryMethod   string     `json:"delivery_method"`
	RecipientName    string     `json:"recipient_name"`
	RecipientContact string     `json:"recipient_contact"`
	Purpose          string     `json:"purpose"`
	Status           string     `json:"status"`
	RequestedAt      time.Time  `json:"requested_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
}

type CreateTranscriptRequestRequest struct {
	DeliveryMethod   string `json:"delivery_method"`
	RecipientName    string `json:"recipient_name"`
	RecipientContact string `json:"recipient_contact"`
	Purpose          string `json:"purpose"`
}

type InternshipProfileResponse struct {
	ID                uint       `json:"id"`
	StudentID         uint       `json:"student_id"`
	ProgrammeID       uint       `json:"programme_id"`
	PreferredIndustry string     `json:"preferred_industry"`
	OrganizationName  string     `json:"organization_name"`
	OrganizationEmail string     `json:"organization_email"`
	OrganizationPhone string     `json:"organization_phone"`
	OrganizationAddr  string     `json:"organization_address"`
	StartDate         *time.Time `json:"start_date,omitempty"`
	EndDate           *time.Time `json:"end_date,omitempty"`
	Status            string     `json:"status"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type UpsertInternshipProfileRequest struct {
	ProgrammeID       uint       `json:"programme_id"`
	PreferredIndustry string     `json:"preferred_industry"`
	OrganizationName  string     `json:"organization_name"`
	OrganizationEmail string     `json:"organization_email"`
	OrganizationPhone string     `json:"organization_phone"`
	OrganizationAddr  string     `json:"organization_address"`
	StartDate         *time.Time `json:"start_date,omitempty"`
	EndDate           *time.Time `json:"end_date,omitempty"`
}
