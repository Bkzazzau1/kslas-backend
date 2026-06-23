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

type StaffRoleAssignmentRequest struct {
StaffID      uint   `json:"staff_id"`
Role        string `json:"role"`
Scope       string `json:"scope"`
FacultyID   *uint  `json:"faculty_id,omitempty"`
DepartmentID *uint `json:"department_id,omitempty"`
ProgrammeID *uint `json:"programme_id,omitempty"`
CourseID     *uint `json:"course_id,omitempty"`
}
