package handlers

import (
	"net/http"

	"kslasbackend/internal/models"
)

func (h *AssessmentHandler) listQuestions(w http.ResponseWriter, r *http.Request) {
	query := h.db.Preload("Options").Preload("Assets").Order("order_number asc")
	if assessmentID := r.URL.Query().Get("assessment_id"); assessmentID != "" {
		query = query.Where("assessment_id = ?", assessmentID)
	}
	var questions []models.Question
	if err := query.Find(&questions).Error; err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, questions)
}

func (h *AssessmentHandler) createQuestion(w http.ResponseWriter, r *http.Request) {
	var question models.Question
	if err := decodeJSON(r, &question); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if question.CourseID.String() == "00000000-0000-0000-0000-000000000000" {
		var assessment models.Assessment
		if err := h.db.First(&assessment, "id = ?", question.AssessmentID).Error; err == nil {
			question.CourseID = assessment.CourseID
		}
	}
	if err := h.db.Create(&question).Error; err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.recalculateAssessmentMarks(question.AssessmentID.String())
	writeJSON(w, http.StatusCreated, question)
}

func (h *AssessmentHandler) createOption(w http.ResponseWriter, r *http.Request) {
	var option models.QuestionOption
	if err := decodeJSON(r, &option); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.db.Create(&option).Error; err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, option)
}

func (h *AssessmentHandler) createAsset(w http.ResponseWriter, r *http.Request) {
	var asset models.QuestionAsset
	if err := decodeJSON(r, &asset); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.db.Create(&asset).Error; err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, asset)
}

func (h *AssessmentHandler) recalculateAssessmentMarks(assessmentID string) {
	var total float64
	h.db.Model(&models.Question{}).Where("assessment_id = ?", assessmentID).Select("COALESCE(SUM(marks), 0)").Scan(&total)
	h.db.Model(&models.Assessment{}).Where("id = ?", assessmentID).Update("total_marks", total)
}
