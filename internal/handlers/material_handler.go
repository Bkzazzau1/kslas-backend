package handlers

import (
	"net/http"
	"strings"

	"kslasbackend/internal/dto"
	"kslasbackend/internal/middleware"
	"kslasbackend/internal/repository"
	"kslasbackend/internal/services"
)

type MaterialHandler struct {
	materialService *services.MaterialService
}

func NewMaterialHandler(materialService *services.MaterialService) *MaterialHandler {
	return &MaterialHandler{materialService: materialService}
}

func (h *MaterialHandler) Materials(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listMaterials(w, r)
	case http.MethodPost:
		h.createMaterial(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (h *MaterialHandler) MaterialByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.getMaterial(w, r)
	case http.MethodPut:
		h.updateMaterial(w, r)
	case http.MethodDelete:
		h.deleteMaterial(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPut, http.MethodDelete)
	}
}

func (h *MaterialHandler) MaterialPublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	userID, materialID, ok := requireUserAndPathID(w, r, "materialID")
	if !ok {
		return
	}
	item, err := h.materialService.PublishMaterial(r.Context(), userID, materialID)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *MaterialHandler) listMaterials(w http.ResponseWriter, r *http.Request) {
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
	includeUnpublished, err := optionalBoolQuery(r, "include_unpublished")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	items, err := h.materialService.ListMaterials(r.Context(), userID, repository.MaterialListFilter{
		CourseID:      courseID,
		CourseCode:    strings.TrimSpace(r.URL.Query().Get("course_code")),
		PublishedOnly: includeUnpublished == nil || !*includeUnpublished,
	})
	if err != nil {
		writeAcademicError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, dto.ListResponse[dto.CourseMaterialResponse]{
		Items: items,
		Count: len(items),
	})
}

func (h *MaterialHandler) createMaterial(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return
	}
	var req dto.CourseMaterialCreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.materialService.CreateMaterial(r.Context(), userID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *MaterialHandler) getMaterial(w http.ResponseWriter, r *http.Request) {
	userID, materialID, ok := requireUserAndPathID(w, r, "materialID")
	if !ok {
		return
	}
	item, err := h.materialService.GetMaterial(r.Context(), userID, materialID)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *MaterialHandler) updateMaterial(w http.ResponseWriter, r *http.Request) {
	userID, materialID, ok := requireUserAndPathID(w, r, "materialID")
	if !ok {
		return
	}
	var req dto.CourseMaterialUpdateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.materialService.UpdateMaterial(r.Context(), userID, materialID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *MaterialHandler) deleteMaterial(w http.ResponseWriter, r *http.Request) {
	userID, materialID, ok := requireUserAndPathID(w, r, "materialID")
	if !ok {
		return
	}
	if err := h.materialService.DeleteMaterial(r.Context(), userID, materialID); err != nil {
		writeAcademicError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
