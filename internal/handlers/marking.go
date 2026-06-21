package handlers

import (
	"encoding/json"
	"strings"

	"github.com/google/uuid"

	"kslasbackend/internal/models"
)

func (h *AssessmentHandler) autoMark(answer *models.StudentAnswer) {
	var question models.Question
	if err := h.db.Preload("Options").First(&question, "id = ?", answer.QuestionID).Error; err != nil {
		return
	}
	answer.Question = question

	if question.RequiresManualMarking || !question.AutoMarkingEnabled {
		answer.MarkingStatus = "needs_review"
		answer.IsAutoMarked = false
		answer.FinalScore = 0
		h.db.Save(answer)
		return
	}

	score := 0.0
	switch question.QuestionType {
	case "single_choice":
		if answer.SelectedOptionID != nil {
			for _, option := range question.Options {
				if option.ID == *answer.SelectedOptionID && option.IsCorrect {
					score = question.Marks
					break
				}
			}
		}
	case "multiple_choice":
		correct := map[uuid.UUID]bool{}
		for _, option := range question.Options {
			if option.IsCorrect {
				correct[option.ID] = true
			}
		}
		var selected []uuid.UUID
		_ = json.Unmarshal(answer.SelectedOptionIDs, &selected)
		if sameUUIDSet(correct, selected) {
			score = question.Marks
		}
	case "fill_blank":
		var metadata struct {
			CorrectAnswers []string `json:"correct_answers"`
		}
		var given []string
		_ = json.Unmarshal(question.Metadata, &metadata)
		_ = json.Unmarshal(answer.BlankAnswers, &given)
		if sameTextList(metadata.CorrectAnswers, given) {
			score = question.Marks
		}
	case "drag_drop":
		if string(question.Metadata) != "" && string(answer.DragDropAnswer) != "" && string(question.Metadata) == string(answer.DragDropAnswer) {
			score = question.Marks
		}
	}

	answer.AutoScore = score
	answer.FinalScore = score
	answer.IsAutoMarked = true
	answer.MarkingStatus = "auto_marked"
	h.db.Save(answer)
	h.recalculateSubmissionScore(answer.SubmissionID.String())
}

func sameUUIDSet(correct map[uuid.UUID]bool, selected []uuid.UUID) bool {
	if len(correct) != len(selected) {
		return false
	}
	for _, id := range selected {
		if !correct[id] {
			return false
		}
	}
	return true
}

func sameTextList(expected []string, given []string) bool {
	if len(expected) != len(given) {
		return false
	}
	for i := range expected {
		if strings.ToLower(strings.TrimSpace(expected[i])) != strings.ToLower(strings.TrimSpace(given[i])) {
			return false
		}
	}
	return true
}

func (h *AssessmentHandler) recalculateSubmissionScore(submissionID string) {
	var total float64
	h.db.Model(&models.StudentAnswer{}).Where("submission_id = ?", submissionID).Select("COALESCE(SUM(final_score), 0)").Scan(&total)
	h.db.Model(&models.StudentSubmission{}).Where("id = ?", submissionID).Update("total_score", total)
}
