package handlers

import (
	"net/http"

	"kslasbackend/internal/dto"
	"kslasbackend/internal/services"
)

type DirectMessageHandler struct {
	messageService *services.DirectMessageService
}

func NewDirectMessageHandler(messageService *services.DirectMessageService) *DirectMessageHandler {
	return &DirectMessageHandler{messageService: messageService}
}

func (h *DirectMessageHandler) CourseMessages(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listCourseMessages(w, r)
	case http.MethodPost:
		h.createCourseMessage(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (h *DirectMessageHandler) listCourseMessages(w http.ResponseWriter, r *http.Request) {
	userID, courseID, ok := requireUserAndPathID(w, r, "courseID")
	if !ok {
		return
	}
	withUserID, err := optionalUintQuery(r, "with_user_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	items, err := h.messageService.ListCourseMessages(r.Context(), userID, courseID, withUserID)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse[dto.CourseDirectMessageResponse]{Items: items, Count: len(items)})
}

func (h *DirectMessageHandler) createCourseMessage(w http.ResponseWriter, r *http.Request) {
	userID, courseID, ok := requireUserAndPathID(w, r, "courseID")
	if !ok {
		return
	}
	var req dto.CourseDirectMessageCreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.messageService.CreateCourseMessage(r.Context(), userID, courseID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}
