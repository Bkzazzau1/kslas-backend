package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"kslasbackend/internal/dto"
	"kslasbackend/internal/middleware"
	"kslasbackend/internal/repository"
	"kslasbackend/internal/services"
)

type AcademicHandler struct {
	academicService *services.AcademicService
}

func NewAcademicHandler(academicService *services.AcademicService) *AcademicHandler {
	return &AcademicHandler{academicService: academicService}
}

func (h *AcademicHandler) Faculties(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listFaculties(w, r)
	case http.MethodPost:
		h.createFaculty(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (h *AcademicHandler) FacultyByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getFaculty(w, r)
	case http.MethodPut:
		h.updateFaculty(w, r)
	case http.MethodDelete:
		h.deleteFaculty(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPut, http.MethodDelete)
	}
}

func (h *AcademicHandler) Departments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listDepartments(w, r)
	case http.MethodPost:
		h.createDepartment(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (h *AcademicHandler) DepartmentByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getDepartment(w, r)
	case http.MethodPut:
		h.updateDepartment(w, r)
	case http.MethodDelete:
		h.deleteDepartment(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPut, http.MethodDelete)
	}
}

func (h *AcademicHandler) Programmes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listProgrammes(w, r)
	case http.MethodPost:
		h.createProgramme(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (h *AcademicHandler) ProgrammeByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getProgramme(w, r)
	case http.MethodPut:
		h.updateProgramme(w, r)
	case http.MethodDelete:
		h.deleteProgramme(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPut, http.MethodDelete)
	}
}

func (h *AcademicHandler) Courses(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listCourses(w, r)
	case http.MethodPost:
		h.createCourse(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (h *AcademicHandler) CourseByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getCourse(w, r)
	case http.MethodPut:
		h.updateCourse(w, r)
	case http.MethodDelete:
		h.deleteCourse(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPut, http.MethodDelete)
	}
}

func (h *AcademicHandler) listFaculties(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return
	}

	items, err := h.academicService.ListFaculties(r.Context(), userID)
	if err != nil {
		writeAcademicError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.ListResponse[dto.FacultyResponse]{
		Items: items,
		Count: len(items),
	})
}

func (h *AcademicHandler) getFaculty(w http.ResponseWriter, r *http.Request) {
	userID, facultyID, ok := requireUserAndPathID(w, r, "facultyID")
	if !ok {
		return
	}

	item, err := h.academicService.GetFaculty(r.Context(), userID, facultyID)
	if err != nil {
		writeAcademicError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h *AcademicHandler) createFaculty(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return
	}

	var req dto.FacultyCreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	item, err := h.academicService.CreateFaculty(r.Context(), userID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func (h *AcademicHandler) updateFaculty(w http.ResponseWriter, r *http.Request) {
	userID, facultyID, ok := requireUserAndPathID(w, r, "facultyID")
	if !ok {
		return
	}

	var req dto.FacultyUpdateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	item, err := h.academicService.UpdateFaculty(r.Context(), userID, facultyID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h *AcademicHandler) deleteFaculty(w http.ResponseWriter, r *http.Request) {
	userID, facultyID, ok := requireUserAndPathID(w, r, "facultyID")
	if !ok {
		return
	}

	if err := h.academicService.DeleteFaculty(r.Context(), userID, facultyID); err != nil {
		writeAcademicError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AcademicHandler) listDepartments(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return
	}

	facultyID, err := optionalUintQuery(r, "faculty_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	items, err := h.academicService.ListDepartments(r.Context(), userID, facultyID)
	if err != nil {
		writeAcademicError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.ListResponse[dto.DepartmentResponse]{
		Items: items,
		Count: len(items),
	})
}

func (h *AcademicHandler) getDepartment(w http.ResponseWriter, r *http.Request) {
	userID, departmentID, ok := requireUserAndPathID(w, r, "departmentID")
	if !ok {
		return
	}

	item, err := h.academicService.GetDepartment(r.Context(), userID, departmentID)
	if err != nil {
		writeAcademicError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h *AcademicHandler) createDepartment(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return
	}

	var req dto.DepartmentCreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	item, err := h.academicService.CreateDepartment(r.Context(), userID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func (h *AcademicHandler) updateDepartment(w http.ResponseWriter, r *http.Request) {
	userID, departmentID, ok := requireUserAndPathID(w, r, "departmentID")
	if !ok {
		return
	}

	var req dto.DepartmentUpdateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	item, err := h.academicService.UpdateDepartment(r.Context(), userID, departmentID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h *AcademicHandler) deleteDepartment(w http.ResponseWriter, r *http.Request) {
	userID, departmentID, ok := requireUserAndPathID(w, r, "departmentID")
	if !ok {
		return
	}

	if err := h.academicService.DeleteDepartment(r.Context(), userID, departmentID); err != nil {
		writeAcademicError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AcademicHandler) listProgrammes(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return
	}

	departmentID, err := optionalUintQuery(r, "department_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	items, err := h.academicService.ListProgrammes(r.Context(), userID, departmentID)
	if err != nil {
		writeAcademicError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.ListResponse[dto.ProgrammeResponse]{
		Items: items,
		Count: len(items),
	})
}

func (h *AcademicHandler) getProgramme(w http.ResponseWriter, r *http.Request) {
	userID, programmeID, ok := requireUserAndPathID(w, r, "programmeID")
	if !ok {
		return
	}

	item, err := h.academicService.GetProgramme(r.Context(), userID, programmeID)
	if err != nil {
		writeAcademicError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h *AcademicHandler) createProgramme(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return
	}

	var req dto.ProgrammeCreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	item, err := h.academicService.CreateProgramme(r.Context(), userID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func (h *AcademicHandler) updateProgramme(w http.ResponseWriter, r *http.Request) {
	userID, programmeID, ok := requireUserAndPathID(w, r, "programmeID")
	if !ok {
		return
	}

	var req dto.ProgrammeUpdateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	item, err := h.academicService.UpdateProgramme(r.Context(), userID, programmeID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h *AcademicHandler) deleteProgramme(w http.ResponseWriter, r *http.Request) {
	userID, programmeID, ok := requireUserAndPathID(w, r, "programmeID")
	if !ok {
		return
	}

	if err := h.academicService.DeleteProgramme(r.Context(), userID, programmeID); err != nil {
		writeAcademicError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AcademicHandler) listCourses(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return
	}

	departmentID, err := optionalUintQuery(r, "department_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	programmeID, err := optionalUintQuery(r, "programme_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	isActive, err := optionalBoolQuery(r, "is_active")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	items, err := h.academicService.ListCourses(r.Context(), userID, repository.CourseListFilter{
		DepartmentID: departmentID,
		ProgrammeID:  programmeID,
		Semester:     strings.TrimSpace(r.URL.Query().Get("semester")),
		Level:        strings.TrimSpace(r.URL.Query().Get("level")),
		IsActive:     isActive,
	})
	if err != nil {
		writeAcademicError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.ListResponse[dto.CourseResponse]{
		Items: items,
		Count: len(items),
	})
}

func (h *AcademicHandler) getCourse(w http.ResponseWriter, r *http.Request) {
	userID, courseID, ok := requireUserAndPathID(w, r, "courseID")
	if !ok {
		return
	}

	item, err := h.academicService.GetCourse(r.Context(), userID, courseID)
	if err != nil {
		writeAcademicError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h *AcademicHandler) createCourse(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return
	}

	var req dto.CourseCreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	item, err := h.academicService.CreateCourse(r.Context(), userID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func (h *AcademicHandler) updateCourse(w http.ResponseWriter, r *http.Request) {
	userID, courseID, ok := requireUserAndPathID(w, r, "courseID")
	if !ok {
		return
	}

	var req dto.CourseUpdateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	item, err := h.academicService.UpdateCourse(r.Context(), userID, courseID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, item)
}

func (h *AcademicHandler) deleteCourse(w http.ResponseWriter, r *http.Request) {
	userID, courseID, ok := requireUserAndPathID(w, r, "courseID")
	if !ok {
		return
	}

	if err := h.academicService.DeleteCourse(r.Context(), userID, courseID); err != nil {
		writeAcademicError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func requireUserAndPathID(w http.ResponseWriter, r *http.Request, pathKey string) (uint, uint, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return 0, 0, false
	}

	value, err := parseUint(r.PathValue(pathKey))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return 0, 0, false
	}

	return userID, value, true
}

func optionalUintQuery(r *http.Request, key string) (*uint, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return nil, nil
	}

	value, err := parseUint(raw)
	if err != nil {
		return nil, errors.New(key + " must be a positive integer")
	}

	return &value, nil
}

func optionalBoolQuery(r *http.Request, key string) (*bool, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return nil, nil
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return nil, errors.New(key + " must be true or false")
	}

	return &value, nil
}

func parseUint(value string) (uint, error) {
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	if err != nil || parsed == 0 {
		return 0, errors.New("invalid identifier")
	}

	return uint(parsed), nil
}

func writeAcademicError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, services.ErrPermissionDenied):
		writeError(w, http.StatusForbidden, "permission denied")
	case errors.Is(err, gorm.ErrRecordNotFound):
		writeError(w, http.StatusNotFound, "resource not found")
	case errors.Is(err, gorm.ErrDuplicatedKey):
		writeError(w, http.StatusConflict, "resource already exists")
	default:
		var validationErr services.ValidationError
		if errors.As(err, &validationErr) {
			writeError(w, http.StatusBadRequest, validationErr.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "request failed")
	}
}

func writeMethodNotAllowed(w http.ResponseWriter, allowed ...string) {
	w.Header().Set("Allow", strings.Join(allowed, ", "))
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}
