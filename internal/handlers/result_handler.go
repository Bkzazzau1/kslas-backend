package handlers

import (
	"net/http"
	"strings"

	"kslasbackend/internal/dto"
	"kslasbackend/internal/middleware"
	"kslasbackend/internal/repository"
	"kslasbackend/internal/services"
)

type ResultHandler struct {
	resultService *services.ResultService
}

func NewResultHandler(resultService *services.ResultService) *ResultHandler {
	return &ResultHandler{resultService: resultService}
}

func (h *ResultHandler) Results(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listResults(w, r)
	case http.MethodPost:
		h.createResult(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (h *ResultHandler) ResultByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getResult(w, r)
	case http.MethodPut:
		h.updateResult(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPut)
	}
}

func (h *ResultHandler) ResultApprove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	userID, resultID, ok := requireUserAndPathID(w, r, "resultID")
	if !ok {
		return
	}
	item, err := h.resultService.ApproveResult(r.Context(), userID, resultID)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ResultHandler) ResultPublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	userID, resultID, ok := requireUserAndPathID(w, r, "resultID")
	if !ok {
		return
	}
	item, err := h.resultService.PublishResult(r.Context(), userID, resultID)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ResultHandler) listResults(w http.ResponseWriter, r *http.Request) {
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
	studentID, err := optionalUintQuery(r, "student_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	items, err := h.resultService.ListResults(r.Context(), userID, repository.ResultListFilter{
		CourseID:       courseID,
		StudentID:      studentID,
		AssessmentType: strings.TrimSpace(r.URL.Query().Get("assessment_type")),
		Status:         strings.TrimSpace(r.URL.Query().Get("status")),
	})
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse[dto.ResultResponse]{
		Items: items,
		Count: len(items),
	})
}

func (h *ResultHandler) getResult(w http.ResponseWriter, r *http.Request) {
	userID, resultID, ok := requireUserAndPathID(w, r, "resultID")
	if !ok {
		return
	}
	item, err := h.resultService.GetResult(r.Context(), userID, resultID)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ResultHandler) createResult(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return
	}
	var req dto.ResultCreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.resultService.CreateResult(r.Context(), userID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *ResultHandler) updateResult(w http.ResponseWriter, r *http.Request) {
	userID, resultID, ok := requireUserAndPathID(w, r, "resultID")
	if !ok {
		return
	}
	var req dto.ResultUpdateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.resultService.UpdateResult(r.Context(), userID, resultID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}
