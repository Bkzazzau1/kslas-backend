package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"kslasbackend/internal/dto"
	"kslasbackend/internal/middleware"
	"kslasbackend/internal/services"
)

type AdministrationHandler struct {
	administrationService *services.AdministrationService
}

func NewAdministrationHandler(administrationService *services.AdministrationService) *AdministrationHandler {
	return &AdministrationHandler{administrationService: administrationService}
}

func (h *AdministrationHandler) Staff(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listStaff(w, r)
	case http.MethodPost:
		h.createStaff(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (h *AdministrationHandler) listStaff(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return
	}
	items, err := h.administrationService.ListStaff(r.Context(), userID)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse[dto.StaffResponse]{Items: items, Count: len(items)})
}

func (h *AdministrationHandler) createStaff(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return
	}
	var req dto.StaffCreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.administrationService.CreateStaff(r.Context(), userID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *AdministrationHandler) StaffPasswordReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	userID, staffID, ok := h.requireStaffAction(w, r)
	if !ok {
		return
	}
	var req dto.StaffResetPasswordRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.administrationService.ResetStaffPassword(r.Context(), userID, staffID, req.Password); err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "password reset successful"})
}

func (h *AdministrationHandler) StaffStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeMethodNotAllowed(w, http.MethodPatch)
		return
	}
	userID, staffID, ok := h.requireStaffAction(w, r)
	if !ok {
		return
	}
	var req dto.StaffStatusUpdateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.administrationService.UpdateStaffStatus(r.Context(), userID, staffID, req.Status)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *AdministrationHandler) requireStaffAction(w http.ResponseWriter, r *http.Request) (uint, uint, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return 0, 0, false
	}
	staffIDRaw := strings.TrimSpace(r.PathValue("staffID"))
	staffID64, err := strconv.ParseUint(staffIDRaw, 10, 64)
	if err != nil || staffID64 == 0 {
		writeError(w, http.StatusBadRequest, "valid staff id is required")
		return 0, 0, false
	}
	return userID, uint(staffID64), true
}

func (h *AdministrationHandler) Students(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return
	}
	var req dto.StudentCreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.administrationService.CreateStudent(r.Context(), userID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *AdministrationHandler) CourseLecturers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	userID, courseID, ok := requireUserAndPathID(w, r, "courseID")
	if !ok {
		return
	}
	var req struct {
		LecturerID uint `json:"lecturer_id"`
	}
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.LecturerID == 0 {
		writeError(w, http.StatusBadRequest, "lecturer_id is required")
		return
	}
	if err := h.administrationService.AssignLecturerToCourse(r.Context(), userID, courseID, req.LecturerID); err != nil {
		writeAcademicError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdministrationHandler) EligibleCourses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}
	studentID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return
	}
	items, err := h.administrationService.ListEligibleCourses(
		r.Context(),
		studentID,
		strings.TrimSpace(r.URL.Query().Get("semester")),
	)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse[dto.CourseResponse]{
		Items: items,
		Count: len(items),
	})
}

func (h *AdministrationHandler) MyCourseRegistrations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listMyCourseRegistrations(w, r)
	case http.MethodPost:
		h.registerMyCourses(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (h *AdministrationHandler) listMyCourseRegistrations(w http.ResponseWriter, r *http.Request) {
	studentID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return
	}
	items, err := h.administrationService.ListMyCourseRegistrations(r.Context(), studentID)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse[dto.CourseRegistrationResponse]{Items: items, Count: len(items)})
}

func (h *AdministrationHandler) registerMyCourses(w http.ResponseWriter, r *http.Request) {
	studentID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return
	}
	var req dto.CourseRegistrationCreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	items, err := h.administrationService.RegisterCourses(r.Context(), studentID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, dto.ListResponse[dto.CourseRegistrationResponse]{Items: items, Count: len(items)})
}

func (h *AdministrationHandler) StaffRoleAssignment(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodPost {
writeMethodNotAllowed(w, http.MethodPost)
return
}

userID, ok := middleware.UserIDFromContext(r.Context())
if !ok {
writeError(w, http.StatusUnauthorized, "unauthenticated user")
return
}

var req dto.StaffRoleAssignmentRequest
if err := decodeJSON(w, r, &req); err != nil {
writeError(w, http.StatusBadRequest, err.Error())
return
}

item, err := h.administrationService.AssignStaffRole(r.Context(), userID, req)
if err != nil {
writeAcademicError(w, err)
return
}

writeJSON(w, http.StatusOK, item)
}
