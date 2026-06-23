package handlers

import (
	"net/http"
	"strings"

	"kslasbackend/internal/dto"
	"kslasbackend/internal/middleware"
	"kslasbackend/internal/repository"
	"kslasbackend/internal/services"
)

type ExamHandler struct {
	examService *services.ExamService
}

func NewExamHandler(examService *services.ExamService) *ExamHandler {
	return &ExamHandler{examService: examService}
}

func (h *ExamHandler) Exams(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listExams(w, r)
	case http.MethodPost:
		h.createExam(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (h *ExamHandler) Venues(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listVenues(w, r)
	case http.MethodPost:
		h.createVenue(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (h *ExamHandler) ExamByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		h.updateExam(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodPut)
	}
}

func (h *ExamHandler) ExamRelease(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	userID, examID, ok := requireUserAndPathID(w, r, "examID")
	if !ok {
		return
	}
	item, err := h.examService.ReleaseExam(r.Context(), userID, examID)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ExamHandler) ExamAttempts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.startAttempt(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodPost)
	}
}

func (h *ExamHandler) ExamStudents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}
	userID, examID, ok := requireUserAndPathID(w, r, "examID")
	if !ok {
		return
	}
	items, err := h.examService.ListRegisteredStudents(r.Context(), userID, examID)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse[dto.ExamStudentResponse]{Items: items, Count: len(items)})
}

func (h *ExamHandler) ExamVenueAllocations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	userID, examID, ok := requireUserAndPathID(w, r, "examID")
	if !ok {
		return
	}
	var req dto.ExamVenueAllocationRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	items, err := h.examService.AllocateVenues(r.Context(), userID, examID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse[dto.ExamStudentAllocationResponse]{Items: items, Count: len(items)})
}

func (h *ExamHandler) startAttempt(w http.ResponseWriter, r *http.Request) {
	userID, examID, ok := requireUserAndPathID(w, r, "examID")
	if !ok {
		return
	}
	var req dto.ExamAttemptStartRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.examService.StartAttempt(r.Context(), userID, examID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *ExamHandler) AttemptSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	userID, attemptID, ok := requireUserAndPathID(w, r, "attemptID")
	if !ok {
		return
	}
	var req dto.ExamAttemptSubmitRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.examService.SubmitAttempt(r.Context(), userID, attemptID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ExamHandler) AttemptShareWithLecturer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	userID, attemptID, ok := requireUserAndPathID(w, r, "attemptID")
	if !ok {
		return
	}
	item, err := h.examService.ShareAttemptWithLecturer(r.Context(), userID, attemptID)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ExamHandler) AttemptMark(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	userID, attemptID, ok := requireUserAndPathID(w, r, "attemptID")
	if !ok {
		return
	}
	var req dto.ExamAttemptMarkRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.examService.MarkAttempt(r.Context(), userID, attemptID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ExamHandler) AttemptModerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	userID, attemptID, ok := requireUserAndPathID(w, r, "attemptID")
	if !ok {
		return
	}
	var req dto.ExamAttemptModerateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.examService.ModerateAttempt(r.Context(), userID, attemptID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ExamHandler) AttemptScriptPDF(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}
	userID, attemptID, ok := requireUserAndPathID(w, r, "attemptID")
	if !ok {
		return
	}
	fileName, data, err := h.examService.ExamScriptPDF(r.Context(), userID, attemptID)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `inline; filename="`+fileName+`"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *ExamHandler) AttemptAnnotations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listAttemptAnnotations(w, r)
	case http.MethodPost:
		h.addAttemptAnnotation(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet, http.MethodPost)
	}
}

func (h *ExamHandler) listAttemptAnnotations(w http.ResponseWriter, r *http.Request) {
	userID, attemptID, ok := requireUserAndPathID(w, r, "attemptID")
	if !ok {
		return
	}
	items, err := h.examService.ListScriptAnnotations(r.Context(), userID, attemptID)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse[dto.ExamScriptAnnotationResponse]{Items: items, Count: len(items)})
}

func (h *ExamHandler) addAttemptAnnotation(w http.ResponseWriter, r *http.Request) {
	userID, attemptID, ok := requireUserAndPathID(w, r, "attemptID")
	if !ok {
		return
	}
	var req dto.ExamScriptAnnotationCreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.examService.AddScriptAnnotation(r.Context(), userID, attemptID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *ExamHandler) AttemptAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	userID, attemptID, ok := requireUserAndPathID(w, r, "attemptID")
	if !ok {
		return
	}
	var req dto.ProctoringAlertCreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.examService.RecordProctoringAlert(r.Context(), userID, attemptID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *ExamHandler) InvigilatorAlerts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.listInvigilatorAlerts(w, r)
	default:
		writeMethodNotAllowed(w, http.MethodGet)
	}
}

func (h *ExamHandler) AlertAcknowledge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	userID, alertID, ok := requireUserAndPathID(w, r, "alertID")
	if !ok {
		return
	}
	item, err := h.examService.AcknowledgeAlert(r.Context(), userID, alertID)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ExamHandler) listExams(w http.ResponseWriter, r *http.Request) {
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
	items, err := h.examService.ListExams(r.Context(), userID, repository.ExamListFilter{
		CourseID: courseID,
		Status:   strings.TrimSpace(r.URL.Query().Get("status")),
	})
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse[dto.ExamResponse]{
		Items: items,
		Count: len(items),
	})
}

func (h *ExamHandler) listVenues(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return
	}
	activeOnly := strings.TrimSpace(r.URL.Query().Get("active")) != "false"
	items, err := h.examService.ListVenues(r.Context(), userID, activeOnly)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse[dto.ExamVenueResponse]{Items: items, Count: len(items)})
}

func (h *ExamHandler) createVenue(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return
	}
	var req dto.ExamVenueCreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.examService.CreateVenue(r.Context(), userID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *ExamHandler) createExam(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return
	}
	var req dto.ExamCreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.examService.CreateExam(r.Context(), userID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *ExamHandler) updateExam(w http.ResponseWriter, r *http.Request) {
	userID, examID, ok := requireUserAndPathID(w, r, "examID")
	if !ok {
		return
	}
	var req dto.ExamUpdateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.examService.UpdateExam(r.Context(), userID, examID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ExamHandler) listInvigilatorAlerts(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated user")
		return
	}
	var acknowledged *bool
	switch strings.TrimSpace(r.URL.Query().Get("acknowledged")) {
	case "true":
		value := true
		acknowledged = &value
	case "false":
		value := false
		acknowledged = &value
	}
	items, err := h.examService.ListInvigilatorAlerts(r.Context(), userID, acknowledged)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, dto.ListResponse[dto.ProctoringAlertResponse]{
		Items: items,
		Count: len(items),
	})
}

func (h *ExamHandler) ExamSubmitToOfficer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	userID, examID, ok := requireUserAndPathID(w, r, "examID")
	if !ok {
		return
	}
	var req dto.ExamWorkflowActionRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.examService.SubmitExamToOfficer(r.Context(), userID, examID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ExamHandler) ExamSendToModerator(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	userID, examID, ok := requireUserAndPathID(w, r, "examID")
	if !ok {
		return
	}
	var req dto.ExamWorkflowActionRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.examService.SendExamToModerator(r.Context(), userID, examID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ExamHandler) ExamModeratorReturn(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	userID, examID, ok := requireUserAndPathID(w, r, "examID")
	if !ok {
		return
	}
	var req dto.ExamWorkflowActionRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.examService.ModeratorReturnExam(r.Context(), userID, examID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ExamHandler) ExamSendBackToLecturer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	userID, examID, ok := requireUserAndPathID(w, r, "examID")
	if !ok {
		return
	}
	var req dto.ExamWorkflowActionRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.examService.SendExamBackToLecturer(r.Context(), userID, examID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *ExamHandler) ExamSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	userID, examID, ok := requireUserAndPathID(w, r, "examID")
	if !ok {
		return
	}
	var req dto.ExamScheduleRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	item, err := h.examService.ScheduleExam(r.Context(), userID, examID, req)
	if err != nil {
		writeAcademicError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}
