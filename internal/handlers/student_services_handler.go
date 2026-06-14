package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"gorm.io/gorm"

	"kslasbackend/internal/dto"
	"kslasbackend/internal/middleware"
	"kslasbackend/internal/services"
)

type StudentServicesHandler struct {
	studentServices *services.StudentServicesService
}

func NewStudentServicesHandler(studentServices *services.StudentServicesService) *StudentServicesHandler {
	return &StudentServicesHandler{studentServices: studentServices}
}

func (h *StudentServicesHandler) GraduationMap(w http.ResponseWriter, r *http.Request) {
	studentID, ok := studentIDFromRequest(w, r)
	if !ok {
		return
	}

	response, err := h.studentServices.GetGraduationMap(r.Context(), studentID)
	if err != nil {
		writeStudentServicesError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *StudentServicesHandler) SupportTickets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listSupportTickets(w, r)
	case http.MethodPost:
		h.createSupportTicket(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (h *StudentServicesHandler) SupportTicketReplies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

	studentID, ok := studentIDFromRequest(w, r)
	if !ok {
		return
	}

	ticketID, ok := pathUint(w, r, "ticketID")
	if !ok {
		return
	}

	var request dto.AddSupportReplyRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	if err := h.studentServices.AddSupportReply(r.Context(), studentID, ticketID, request); err != nil {
		writeStudentServicesError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"message": "reply added"})
}

func (h *StudentServicesHandler) InternshipProfile(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getInternshipProfile(w, r)
	case http.MethodPut:
		h.upsertInternshipProfile(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPut)
	}
}

func (h *StudentServicesHandler) TranscriptRequests(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listTranscriptRequests(w, r)
	case http.MethodPost:
		h.createTranscriptRequest(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (h *StudentServicesHandler) listSupportTickets(w http.ResponseWriter, r *http.Request) {
	studentID, ok := studentIDFromRequest(w, r)
	if !ok {
		return
	}

	items, err := h.studentServices.ListSupportTickets(r.Context(), studentID)
	if err != nil {
		writeStudentServicesError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.ListResponse[dto.SupportTicketResponse]{Items: items, Count: len(items)})
}

func (h *StudentServicesHandler) createSupportTicket(w http.ResponseWriter, r *http.Request) {
	studentID, ok := studentIDFromRequest(w, r)
	if !ok {
		return
	}

	var request dto.CreateSupportTicketRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	response, err := h.studentServices.CreateSupportTicket(r.Context(), studentID, request)
	if err != nil {
		writeStudentServicesError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, response)
}

func (h *StudentServicesHandler) getInternshipProfile(w http.ResponseWriter, r *http.Request) {
	studentID, ok := studentIDFromRequest(w, r)
	if !ok {
		return
	}

	response, err := h.studentServices.GetInternshipProfile(r.Context(), studentID)
	if err != nil {
		writeStudentServicesError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *StudentServicesHandler) upsertInternshipProfile(w http.ResponseWriter, r *http.Request) {
	studentID, ok := studentIDFromRequest(w, r)
	if !ok {
		return
	}

	var request dto.UpsertInternshipProfileRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	response, err := h.studentServices.UpsertInternshipProfile(r.Context(), studentID, request)
	if err != nil {
		writeStudentServicesError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (h *StudentServicesHandler) listTranscriptRequests(w http.ResponseWriter, r *http.Request) {
	studentID, ok := studentIDFromRequest(w, r)
	if !ok {
		return
	}

	items, err := h.studentServices.ListTranscriptRequests(r.Context(), studentID)
	if err != nil {
		writeStudentServicesError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.ListResponse[dto.TranscriptRequestResponse]{Items: items, Count: len(items)})
}

func (h *StudentServicesHandler) createTranscriptRequest(w http.ResponseWriter, r *http.Request) {
	studentID, ok := studentIDFromRequest(w, r)
	if !ok {
		return
	}

	var request dto.CreateTranscriptRequestRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	response, err := h.studentServices.CreateTranscriptRequest(r.Context(), studentID, request)
	if err != nil {
		writeStudentServicesError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, response)
}

func studentIDFromRequest(w http.ResponseWriter, r *http.Request) (uint, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == 0 {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return 0, false
	}
	return userID, true
}

func pathUint(w http.ResponseWriter, r *http.Request, name string) (uint, bool) {
	value := r.PathValue(name)
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		writeError(w, http.StatusBadRequest, "invalid "+name)
		return 0, false
	}
	return uint(id), true
}

func writeStudentServicesError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		writeError(w, http.StatusNotFound, "record not found")
	default:
		writeError(w, http.StatusBadRequest, err.Error())
	}
}
