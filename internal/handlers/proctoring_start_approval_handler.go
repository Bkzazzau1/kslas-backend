package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"kslasbackend/internal/dto"
)

func (h *ProctoringReviewHandler) StartApproval(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

	var req dto.ExamStartApprovalRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if strings.TrimSpace(req.StudentID) == "" || strings.TrimSpace(req.ExamID) == "" || strings.TrimSpace(req.AttemptID) == "" {
		writeError(w, http.StatusBadRequest, "student_id, exam_id, and attempt_id are required")
		return
	}

	issues := []string{}
	if !req.FaceIDReady {
		issues = append(issues, "Face ID is not ready.")
	}
	if !req.RoomScanReady || strings.TrimSpace(req.ManifestPath) == "" {
		issues = append(issues, "360 room scan evidence is not ready.")
	}
	if !req.AudioReady {
		issues = append(issues, "Audio review is not ready.")
	}
	if !req.SystemReady {
		issues = append(issues, "System device review is not ready.")
	}

	blockingSystemIssue := boolFromReview(req.SystemReview, "bluetooth_detected") ||
		boolFromReview(req.SystemReview, "external_audio_detected") ||
		boolFromReview(req.SystemReview, "usb_risk_detected") ||
		boolFromReview(req.SystemReview, "virtualization_detected") ||
		boolFromReview(req.SystemReview, "container_detected") ||
		boolFromReview(req.SystemReview, "virtual_camera_detected") ||
		boolFromReview(req.SystemReview, "unknown_device_state")

	if blockingSystemIssue {
		issues = append(issues, "Blocking system review issue is still present.")
	}

	warningOnly := boolFromReview(req.SystemReview, "virtualization_warning_detected") && !blockingSystemIssue

	response := dto.ExamStartApprovalResponse{
		Status:              "approved_to_start",
		ApprovalSource:      "backend_rules",
		AIRecommendation:    "low_risk",
		RequiresHumanReview: false,
		Message:             "Backend approved this attempt. The exam may start.",
		Issues:              issues,
	}

	if len(issues) > 0 {
		response.Status = "blocked"
		response.AIRecommendation = "high_risk"
		response.RequiresHumanReview = true
		response.Message = "Start approval denied. Resolve blocking setup issues and request approval again."
		writeJSON(w, http.StatusOK, response)
		return
	}

	if warningOnly {
		response.AIRecommendation = "review_note"
		response.Message = "Backend approved this attempt with a recorded Windows security virtualization note."
	}

	token, err := newExamStartToken(req.AttemptID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not generate exam start token")
		return
	}
	response.ExamStartToken = token
	response.ExpiresAt = time.Now().UTC().Add(30 * time.Minute).Format(time.RFC3339)
	writeJSON(w, http.StatusOK, response)
}

func boolFromReview(review map[string]interface{}, key string) bool {
	if review == nil {
		return false
	}
	value, ok := review[key]
	if !ok {
		return false
	}
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return strings.EqualFold(strings.TrimSpace(typed), "true")
	default:
		return fmt.Sprint(typed) == "true"
	}
}

func newExamStartToken(attemptID string) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "exam_start_" + strings.TrimSpace(attemptID) + "_" + base64.RawURLEncoding.EncodeToString(bytes), nil
}
