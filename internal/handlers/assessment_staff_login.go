package handlers

import (
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"kslasbackend/internal/middleware"
	"kslasbackend/internal/models"
)

type assessmentStaffLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type assessmentStaffLoginResponse struct {
	Token string       `json:"token"`
	Staff models.Staff `json:"staff"`
	Roles []string     `json:"roles"`
}

func (h *AssessmentHandler) staffLogin(w http.ResponseWriter, r *http.Request) {
	var payload assessmentStaffLoginRequest
	if err := decodeJSON(w, r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	payload.Email = strings.TrimSpace(strings.ToLower(payload.Email))
	if payload.Email == "" || payload.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	var staff models.Staff
	if err := h.db.Preload("Department").Where("LOWER(email) = ?", payload.Email).First(&staff).Error; err != nil {
		writeError(w, http.StatusUnauthorized, "invalid login details")
		return
	}
	if !staff.IsActive {
		writeError(w, http.StatusForbidden, "staff account is not active")
		return
	}
	if staff.PasswordHash == "" || bcrypt.CompareHashAndPassword([]byte(staff.PasswordHash), []byte(payload.Password)) != nil {
		writeError(w, http.StatusUnauthorized, "invalid login details")
		return
	}

	roles := h.staffRoleNames(staff)
	token, err := middleware.NewStaffToken(staff.ID, staff.PrimaryRole, roles)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	now := nowPtr()
	h.db.Model(&staff).Update("last_login_at", now)
	staff.LastLoginAt = now

	writeJSON(w, http.StatusOK, assessmentStaffLoginResponse{Token: token, Staff: staff, Roles: roles})
}

func (h *AssessmentHandler) staffRoleNames(staff models.Staff) []string {
	roles := []string{}
	if staff.PrimaryRole != "" {
		roles = append(roles, staff.PrimaryRole)
	}
	var assignments []models.StaffRoleAssignment
	if err := h.db.Where("staff_id = ? AND is_active = ?", staff.ID, true).Find(&assignments).Error; err == nil {
		for _, assignment := range assignments {
			if assignment.Role != "" {
				roles = append(roles, assignment.Role)
			}
		}
	}
	seen := map[string]bool{}
	unique := []string{}
	for _, role := range roles {
		if !seen[role] {
			seen[role] = true
			unique = append(unique, role)
		}
	}
	return unique
}
