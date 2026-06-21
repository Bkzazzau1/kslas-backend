package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"kslasbackend/internal/models"
)

func (h *AssessmentHandler) listSubmissions(w http.ResponseWriter, r *http.Request) {
	query := h.db.Preload("Answers").Order("created_at desc")
	if assessmentID := r.URL.Query().Get("assessment_id"); assessmentID != "" {
		query = query.Where("assessment_id = ?", assessmentID)
	}
	var submissions []models.StudentSubmission
	if err := query.Find(&submissions).Error; err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, submissions)
}

func (h *AssessmentHandler) listPublishedAssessments(w http.ResponseWriter, r *http.Request) {
	var assessments []models.Assessment
	if err := h.db.Preload("Course").Preload("Questions.Options").Preload("Questions.Assets").Where("status IN ?", []string{"published", "active"}).Find(&assessments).Error; err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, assessments)
}

func (h *AssessmentHandler) studentAssessmentAction(w http.ResponseWriter, r *http.Request) {
	id, action, ok := splitIDAction(r.URL.Path, "/api/student/assessments/")
	if !ok {
		writeError(w, http.StatusNotFound, "invalid student assessment action")
		return
	}

	studentID, err := uuid.Parse(r.URL.Query().Get("student_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "student_id query parameter is required")
		return
	}

	switch action {
	case "start":
		submission := models.StudentSubmission{AssessmentID: id, StudentID: studentID}
		if err := h.db.FirstOrCreate(&submission, models.StudentSubmission{AssessmentID: id, StudentID: studentID}).Error; err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, submission)
	case "submit":
		var submission models.StudentSubmission
		if err := h.db.Where("assessment_id = ? AND student_id = ?", id, studentID).First(&submission).Error; err != nil {
			writeError(w, http.StatusNotFound, "submission not found")
			return
		}
		submission.Status = "submitted"
		submission.SubmittedAt = nowPtr()
		h.db.Save(&submission)
		writeJSON(w, http.StatusOK, submission)
	default:
		writeError(w, http.StatusNotFound, "unknown action")
	}
}

func (h *AssessmentHandler) submitAnswer(w http.ResponseWriter, r *http.Request) {
	var answer models.StudentAnswer
	if err := decodeJSON(r, &answer); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.db.Where("submission_id = ? AND question_id = ?", answer.SubmissionID, answer.QuestionID).Assign(answer).FirstOrCreate(&answer).Error; err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.autoMark(&answer)
	writeJSON(w, http.StatusCreated, answer)
}

func (h *AssessmentHandler) markAnswer(w http.ResponseWriter, r *http.Request) {
	idText := strings.TrimPrefix(r.URL.Path, "/api/lecturer/answers/")
	idText = strings.TrimSuffix(idText, "/mark")
	idText = strings.Trim(idText, "/")
	answerID, err := uuid.Parse(idText)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid answer id")
		return
	}

	var payload struct {
		ManualScore      float64 `json:"manual_score"`
		LecturerFeedback string  `json:"lecturer_feedback"`
		MarkedByID       *uuid.UUID `json:"marked_by_id"`
	}
	if err := decodeJSON(r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var answer models.StudentAnswer
	if err := h.db.Preload("Question").First(&answer, "id = ?", answerID).Error; err != nil {
		writeError(w, http.StatusNotFound, "answer not found")
		return
	}
	if payload.ManualScore > answer.Question.Marks {
		writeError(w, http.StatusBadRequest, "score cannot be higher than question mark")
		return
	}
	answer.ManualScore = &payload.ManualScore
	answer.FinalScore = payload.ManualScore
	answer.LecturerFeedback = payload.LecturerFeedback
	answer.MarkingStatus = "marked"
	answer.MarkedByID = payload.MarkedByID
	answer.MarkedAt = nowPtr()
	if err := h.db.Save(&answer).Error; err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.recalculateSubmissionScore(answer.SubmissionID.String())
	writeJSON(w, http.StatusOK, answer)
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
