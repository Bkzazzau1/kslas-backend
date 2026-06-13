package handlers

import (
	"net/http"
	"strings"

	"kslasbackend/internal/dto"
	"kslasbackend/internal/middleware"
	"kslasbackend/internal/repository"
	"kslasbackend/internal/services"
)

type AssignmentHandler struct {
	assignmentService *services.AssignmentService
}

func NewAssignmentHandler(assignmentService *services.AssignmentService) *AssignmentHandler {
	return &AssignmentHandler{assignmentService: assignmentService}
}

func (h *AssignmentHandler) Assignments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listAssignments(w, r)
	case http.MethodPost:
		h.createAssignment(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (h *AssignmentHandler) AssignmentByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getAssignment(w, r)
	case http.MethodPut:
		h.updateAssignment(w, r)
	case http.MethodDelete:
		h.deleteAssignment(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPut, http.MethodDelete)
	}
}

func (h *AssignmentHandler) AssignmentSubmissions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listSubmissions(w, r)
	case http.MethodPost:
		h.submitAssignment(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (h *AssignmentHandler) AssignmentGrades(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listGrades(w, r)
	case http.MethodPost:
		h.upsertGrades(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (h *AssignmentHandler) AssignmentPeerReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	h.submitPeerReview(w, r)
}

func (h *AssignmentHandler) listAssignments(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return
	}

	courseID, err := optionalUintQuery(r, "course_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	items, err := h.assignmentService.ListAssignments(r.Context(), userID, repository.AssignmentListFilter{
		CourseID: courseID,
		Status:   strings.TrimSpace(r.URL.Query().Get("status")),
	})
	if err != nil {
		writeAcademicError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.ListResponse[dto.AssignmentResponse]{
		Items: items,
		Count: len(items),
	})
}

func (h *AssignmentHandler) getAssignment(w http.ResponseWriter, r *http.Request) {
	userID, assignmentID, ok := requireUserAndPathID(w, r, "assignmentID")
	if !ok {
		return
	}

	item, err := h.assignmentService.GetAssignment(r.Context(), userID, assignmentID)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *AssignmentHandler) createAssignment(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return
	}

	var req dto.AssignmentCreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.assignmentService.CreateAssignment(r.Context(), userID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *AssignmentHandler) updateAssignment(w http.ResponseWriter, r *http.Request) {
	userID, assignmentID, ok := requireUserAndPathID(w, r, "assignmentID")
	if !ok {
		return
	}

	var req dto.AssignmentUpdateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.assignmentService.UpdateAssignment(r.Context(), userID, assignmentID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *AssignmentHandler) deleteAssignment(w http.ResponseWriter, r *http.Request) {
	userID, assignmentID, ok := requireUserAndPathID(w, r, "assignmentID")
	if !ok {
		return
	}

	if err := h.assignmentService.DeleteAssignment(r.Context(), userID, assignmentID); err != nil {
		writeAcademicError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AssignmentHandler) submitAssignment(w http.ResponseWriter, r *http.Request) {
	userID, assignmentID, ok := requireUserAndPathID(w, r, "assignmentID")
	if !ok {
		return
	}

	var req dto.AssignmentSubmissionCreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.assignmentService.SubmitAssignment(r.Context(), userID, assignmentID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *AssignmentHandler) listSubmissions(w http.ResponseWriter, r *http.Request) {
	userID, assignmentID, ok := requireUserAndPathID(w, r, "assignmentID")
	if !ok {
		return
	}

	items, err := h.assignmentService.ListSubmissions(r.Context(), userID, assignmentID)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse[dto.AssignmentSubmissionResponse]{
		Items: items,
		Count: len(items),
	})
}

func (h *AssignmentHandler) submitPeerReview(w http.ResponseWriter, r *http.Request) {
	userID, assignmentID, ok := requireUserAndPathID(w, r, "assignmentID")
	if !ok {
		return
	}
	reviewID, err := parseUint(r.PathValue("reviewID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req dto.AssignmentPeerReviewSubmitRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.assignmentService.SubmitPeerReview(r.Context(), userID, assignmentID, reviewID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *AssignmentHandler) listGrades(w http.ResponseWriter, r *http.Request) {
	userID, assignmentID, ok := requireUserAndPathID(w, r, "assignmentID")
	if !ok {
		return
	}

	items, err := h.assignmentService.ListGrades(r.Context(), userID, assignmentID)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse[dto.AssignmentGradeResponse]{
		Items: items,
		Count: len(items),
	})
}

func (h *AssignmentHandler) upsertGrades(w http.ResponseWriter, r *http.Request) {
	userID, assignmentID, ok := requireUserAndPathID(w, r, "assignmentID")
	if !ok {
		return
	}

	var req dto.AssignmentGradeBulkUpsertRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	items, err := h.assignmentService.UpsertGrades(r.Context(), userID, assignmentID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse[dto.AssignmentGradeResponse]{
		Items: items,
		Count: len(items),
	})
}
