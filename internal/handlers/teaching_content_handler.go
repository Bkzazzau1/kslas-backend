package handlers

import (
	"net/http"
	"strings"

	"kslasbackend/internal/dto"
	"kslasbackend/internal/middleware"
	"kslasbackend/internal/repository"
	"kslasbackend/internal/services"
)

type TeachingContentHandler struct {
	contentService *services.TeachingContentService
}

func NewTeachingContentHandler(contentService *services.TeachingContentService) *TeachingContentHandler {
	return &TeachingContentHandler{contentService: contentService}
}

func (h *TeachingContentHandler) VideoLectures(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listVideoLectures(w, r)
	case http.MethodPost:
		h.createVideoLecture(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (h *TeachingContentHandler) VideoLectureByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeMethodNotAllowed(w, http.MethodDelete)
		return
	}
	userID, lectureID, ok := requireUserAndPathID(w, r, "lectureID")
	if !ok {
		return
	}
	if err := h.contentService.DeleteVideoLecture(r.Context(), userID, lectureID); err != nil {
		writeAcademicError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TeachingContentHandler) VideoLectureWatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	userID, lectureID, ok := requireUserAndPathID(w, r, "lectureID")
	if !ok {
		return
	}
	var req dto.VideoLectureWatchRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.contentService.MarkVideoLectureWatched(r.Context(), userID, lectureID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *TeachingContentHandler) LiveSessions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listLiveSessions(w, r)
	case http.MethodPost:
		h.createLiveSession(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (h *TeachingContentHandler) LiveSessionByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeMethodNotAllowed(w, http.MethodPut)
		return
	}
	userID, sessionID, ok := requireUserAndPathID(w, r, "sessionID")
	if !ok {
		return
	}
	var req dto.LiveSessionUpdateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.contentService.UpdateLiveSession(r.Context(), userID, sessionID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *TeachingContentHandler) listVideoLectures(w http.ResponseWriter, r *http.Request) {
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
	items, err := h.contentService.ListVideoLectures(r.Context(), userID, repository.VideoLectureListFilter{CourseID: courseID})
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse[dto.VideoLectureResponse]{Items: items, Count: len(items)})
}

func (h *TeachingContentHandler) createVideoLecture(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return
	}
	var req dto.VideoLectureCreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.contentService.CreateVideoLecture(r.Context(), userID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *TeachingContentHandler) listLiveSessions(w http.ResponseWriter, r *http.Request) {
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
	items, err := h.contentService.ListLiveSessions(r.Context(), userID, repository.LiveSessionListFilter{
		CourseID: courseID,
		Status:   strings.TrimSpace(r.URL.Query().Get("status")),
	})
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse[dto.LiveSessionResponse]{Items: items, Count: len(items)})
}

func (h *TeachingContentHandler) createLiveSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return
	}
	var req dto.LiveSessionCreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.contentService.CreateLiveSession(r.Context(), userID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}
