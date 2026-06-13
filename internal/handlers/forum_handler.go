package handlers

import (
	"net/http"

	"kslasbackend/internal/dto"
	"kslasbackend/internal/middleware"
	"kslasbackend/internal/services"
)

type ForumHandler struct {
	forumService *services.ForumService
}

func NewForumHandler(forumService *services.ForumService) *ForumHandler {
	return &ForumHandler{forumService: forumService}
}

func (h *ForumHandler) CourseForumPosts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listCoursePosts(w, r)
	case http.MethodPost:
		h.createCoursePost(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (h *ForumHandler) CourseForumPostByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeMethodNotAllowed(w, http.MethodPatch)
		return
	}
	h.moderateCoursePost(w, r)
}

func (h *ForumHandler) listCoursePosts(w http.ResponseWriter, r *http.Request) {
	userID, courseID, ok := h.requireUserAndCourse(w, r)
	if !ok {
		return
	}
	items, err := h.forumService.ListCoursePosts(r.Context(), userID, courseID)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse[dto.CourseForumPostResponse]{
		Items: items,
		Count: len(items),
	})
}

func (h *ForumHandler) createCoursePost(w http.ResponseWriter, r *http.Request) {
	userID, courseID, ok := h.requireUserAndCourse(w, r)
	if !ok {
		return
	}
	var req dto.CourseForumPostCreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.forumService.CreateCoursePost(r.Context(), userID, courseID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *ForumHandler) moderateCoursePost(w http.ResponseWriter, r *http.Request) {
	userID, courseID, ok := h.requireUserAndCourse(w, r)
	if !ok {
		return
	}
	postID, err := parseUint(r.PathValue("postID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var req dto.CourseForumPostModerationRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.forumService.ModerateCoursePost(r.Context(), userID, courseID, postID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ForumHandler) requireUserAndCourse(w http.ResponseWriter, r *http.Request) (uint, uint, bool) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return 0, 0, false
	}
	courseID, err := parseUint(r.PathValue("courseID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return 0, 0, false
	}
	return userID, courseID, true
}
