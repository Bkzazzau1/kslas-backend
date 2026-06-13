package dto

import "time"

type LoginRequest struct {
	Identity string `json:"identity"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken      string      `json:"access_token"`
	TokenType        string      `json:"token_type"`
	ExpiresInSeconds int         `json:"expires_in_seconds"`
	User             UserPayload `json:"user"`
}

type UserPayload struct {
	ID          uint              `json:"id"`
	UUID        string            `json:"uuid"`
	FirstName   string            `json:"first_name"`
	LastName    string            `json:"last_name"`
	MiddleName  string            `json:"middle_name,omitempty"`
	Email       string            `json:"email,omitempty"`
	Phone       string            `json:"phone,omitempty"`
	UserType    string            `json:"user_type"`
	Status      string            `json:"status"`
	LastLoginAt *time.Time        `json:"last_login_at,omitempty"`
	Roles       []UserRolePayload `json:"roles"`
}

type UserRolePayload struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	ScopeType string `json:"scope_type"`
	ScopeID   *uint  `json:"scope_id,omitempty"`
	IsPrimary bool   `json:"is_primary"`
}
