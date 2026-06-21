package handlers

import (
	"net/http"

	"github.com/google/uuid"

	"kslasbackend/internal/middleware"
	"kslasbackend/internal/models"
)

func (h *AssessmentHandler) listStaff(w http.ResponseWriter, r *http.Request) {
	if !h.requireAnyRole(w, r, "admin", "dlc_director", "hod") {
		return
	}
	query := h.db.Preload("Department").Order("created_at desc")
	if role := r.URL.Query().Get("role"); role != "" {
		query = query.Where("primary_role = ?", role)
	}
	if departmentID := r.URL.Query().Get("department_id"); departmentID != "" {
		query = query.Where("department_id = ?", departmentID)
	}
	var staff []models.Staff
	if err := query.Find(&staff).Error; err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, staff)
}

func (h *AssessmentHandler) createStaff(w http.ResponseWriter, r *http.Request) {
	if !h.requireAnyRole(w, r, "admin", "dlc_director", "hod") {
		return
	}
	var staff models.Staff
	if err := decodeJSON(w, r, &staff); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.db.Create(&staff).Error; err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, staff)
}

func (h *AssessmentHandler) getMyStaffProfile(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.StaffClaimsFromHeaders(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "staff authentication headers are required")
		return
	}
	var staff models.Staff
	if err := h.db.Preload("Department").First(&staff, "id = ?", claims.ID).Error; err != nil {
		writeError(w, http.StatusNotFound, "staff profile not found")
		return
	}
	writeJSON(w, http.StatusOK, staff)
}

func (h *AssessmentHandler) listStaffRoles(w http.ResponseWriter, r *http.Request) {
	if !h.requireAnyRole(w, r, "admin", "dlc_director", "hod") {
		return
	}
	query := h.db.Preload("Staff").Order("created_at desc")
	if staffID := r.URL.Query().Get("staff_id"); staffID != "" {
		query = query.Where("staff_id = ?", staffID)
	}
	if role := r.URL.Query().Get("role"); role != "" {
		query = query.Where("role = ?", role)
	}
	var roles []models.StaffRoleAssignment
	if err := query.Find(&roles).Error; err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, roles)
}

func (h *AssessmentHandler) createStaffRole(w http.ResponseWriter, r *http.Request) {
	if !h.requireAnyRole(w, r, "admin", "dlc_director", "hod") {
		return
	}
	var role models.StaffRoleAssignment
	if err := decodeJSON(w, r, &role); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if role.StaffID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "staff_id is required")
		return
	}
	if role.Role == "" {
		writeError(w, http.StatusBadRequest, "role is required")
		return
	}
	claims, ok := middleware.StaffClaimsFromHeaders(r)
	if ok && role.AssignedByID == nil {
		role.AssignedByID = &claims.ID
	}
	if err := h.db.Create(&role).Error; err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, role)
}

func (h *AssessmentHandler) requireAnyRole(w http.ResponseWriter, r *http.Request, roles ...string) bool {
	claims, ok := middleware.StaffClaimsFromHeaders(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "staff authentication headers are required")
		return false
	}
	allowed := map[string]bool{}
	for _, role := range roles {
		allowed[role] = true
	}
	if !claims.HasAnyRole(allowed) {
		writeError(w, http.StatusForbidden, "staff role is not allowed for this action")
		return false
	}
	return true
}
