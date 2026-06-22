package dto

import "time"

type StaffResponse struct {
	ID              uint              `json:"id"`
	UUID            string            `json:"uuid"`
	FirstName       string            `json:"first_name"`
	LastName        string            `json:"last_name"`
	MiddleName      string            `json:"middle_name,omitempty"`
	Email           string            `json:"email,omitempty"`
	Phone           string            `json:"phone,omitempty"`
	Gender          string            `json:"gender,omitempty"`
	MatricNo        string            `json:"matric_no,omitempty"`
	StaffID         string            `json:"staff_id,omitempty"`
	UserType        string            `json:"user_type"`
	Status          string            `json:"status"`
	PrimaryRoleCode string            `json:"primary_role,omitempty"`
	Roles           []UserRolePayload `json:"roles"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

type StaffResetPasswordRequest struct {
	Password string `json:"password"`
}

type StaffStatusUpdateRequest struct {
	Status string `json:"status"`
}
