package handlers

import (
	"net/http"

	"kslasbackend/internal/dto"
	"kslasbackend/internal/middleware"
	"kslasbackend/internal/services"
)

type ReportHandler struct {
	reportService *services.ReportService
}

func NewReportHandler(reportService *services.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

func (h *ReportHandler) LecturerReports(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}
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
	lecturerID, err := optionalUintQuery(r, "lecturer_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	items, err := h.reportService.LecturerReports(r.Context(), userID, courseID, lecturerID)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse[dto.LecturerCourseReportResponse]{Items: items, Count: len(items)})
}
