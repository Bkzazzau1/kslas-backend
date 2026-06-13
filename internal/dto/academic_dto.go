package dto

import "time"

type FacultyCreateRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type FacultyUpdateRequest struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type FacultyResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DepartmentCreateRequest struct {
	FacultyID uint   `json:"faculty_id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
}

type DepartmentUpdateRequest struct {
	FacultyID uint   `json:"faculty_id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
}

type DepartmentResponse struct {
	ID        uint      `json:"id"`
	FacultyID uint      `json:"faculty_id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProgrammeCreateRequest struct {
	DepartmentID uint   `json:"department_id"`
	Name         string `json:"name"`
	Code         string `json:"code"`
	LevelType    string `json:"level_type"`
}

type ProgrammeUpdateRequest struct {
	DepartmentID uint   `json:"department_id"`
	Name         string `json:"name"`
	Code         string `json:"code"`
	LevelType    string `json:"level_type"`
}

type ProgrammeResponse struct {
	ID           uint      `json:"id"`
	DepartmentID uint      `json:"department_id"`
	Name         string    `json:"name"`
	Code         string    `json:"code"`
	LevelType    string    `json:"level_type"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CourseCreateRequest struct {
	DepartmentID uint   `json:"department_id"`
	ProgrammeID  *uint  `json:"programme_id"`
	Title        string `json:"title"`
	Code         string `json:"code"`
	Unit         uint   `json:"unit"`
	Semester     string `json:"semester"`
	Level        string `json:"level"`
	IsActive     *bool  `json:"is_active"`
}

type CourseUpdateRequest struct {
	DepartmentID uint   `json:"department_id"`
	ProgrammeID  *uint  `json:"programme_id"`
	Title        string `json:"title"`
	Code         string `json:"code"`
	Unit         uint   `json:"unit"`
	Semester     string `json:"semester"`
	Level        string `json:"level"`
	IsActive     *bool  `json:"is_active"`
}

type CourseResponse struct {
	ID           uint      `json:"id"`
	DepartmentID uint      `json:"department_id"`
	ProgrammeID  *uint     `json:"programme_id,omitempty"`
	Title        string    `json:"title"`
	Code         string    `json:"code"`
	Unit         uint      `json:"unit"`
	Semester     string    `json:"semester"`
	Level        string    `json:"level"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type StaffCreateRequest struct {
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	MiddleName   string `json:"middle_name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Password     string `json:"password"`
	Gender       string `json:"gender"`
	StaffID      string `json:"staff_id"`
	RoleCode     string `json:"role_code"`
	DepartmentID uint   `json:"department_id"`
}

type StudentCreateRequest struct {
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	MiddleName      string `json:"middle_name"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Password        string `json:"password"`
	Gender          string `json:"gender"`
	MatricNo        string `json:"matric_no"`
	DepartmentID    uint   `json:"department_id"`
	ProgrammeID     uint   `json:"programme_id"`
	Level           string `json:"level"`
	Semester        string `json:"semester"`
	AcademicSession string `json:"academic_session"`
}

type UserResponse struct {
	ID         uint      `json:"id"`
	UUID       string    `json:"uuid"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	MiddleName string    `json:"middle_name,omitempty"`
	Email      string    `json:"email,omitempty"`
	Phone      string    `json:"phone,omitempty"`
	Gender     string    `json:"gender,omitempty"`
	MatricNo   string    `json:"matric_no,omitempty"`
	StaffID    string    `json:"staff_id,omitempty"`
	UserType   string    `json:"user_type"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type StudentAcademicProfileResponse struct {
	ID                 uint      `json:"id"`
	StudentID          uint      `json:"student_id"`
	DepartmentID       uint      `json:"department_id"`
	ProgrammeID        uint      `json:"programme_id"`
	Level              string    `json:"level"`
	Semester           string    `json:"semester"`
	AcademicSession    string    `json:"academic_session"`
	CountryOfResidence string    `json:"country_of_residence,omitempty"`
	IsInternational    bool      `json:"is_international"`
	RegisteredBy       uint      `json:"registered_by"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type StudentRegistrationResponse struct {
	User    UserResponse                   `json:"user"`
	Profile StudentAcademicProfileResponse `json:"profile"`
}

type CourseRegistrationCreateRequest struct {
	CourseIDs       []uint `json:"course_ids"`
	AcademicSession string `json:"academic_session"`
	Semester        string `json:"semester"`
}

type CourseRegistrationResponse struct {
	ID              uint           `json:"id"`
	StudentID       uint           `json:"student_id"`
	CourseID        uint           `json:"course_id"`
	CourseCode      string         `json:"course_code,omitempty"`
	CourseTitle     string         `json:"course_title,omitempty"`
	AcademicSession string         `json:"academic_session"`
	Semester        string         `json:"semester"`
	Level           string         `json:"level"`
	Status          string         `json:"status"`
	RegisteredAt    time.Time      `json:"registered_at"`
	Course          CourseResponse `json:"course,omitempty"`
}

type ListResponse[T any] struct {
	Items []T `json:"items"`
	Count int `json:"count"`
}
