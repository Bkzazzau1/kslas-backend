package handlers

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"kslasbackend/internal/dto"
)

type ProctoringReviewHandler struct{}

func NewProctoringReviewHandler() *ProctoringReviewHandler {
	return &ProctoringReviewHandler{}
}

func (h *ProctoringReviewHandler) PreExamReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

	if err := r.ParseMultipartForm(128 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid evidence record")
		return
	}

	manifest := strings.TrimSpace(r.FormValue("manifest"))
	if manifest == "" {
		writeError(w, http.StatusBadRequest, "manifest is required")
		return
	}

	var req dto.PreExamReviewRequest
	if err := json.Unmarshal([]byte(manifest), &req); err != nil {
		writeError(w, http.StatusBadRequest, "manifest must be valid json")
		return
	}

	writeJSON(w, http.StatusOK, fallbackPreExamReview(req, uploadedEvidenceFiles(r.MultipartForm)))
}

func (h *ProctoringReviewHandler) LiveEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

	var req dto.LiveProctoringEventRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if strings.TrimSpace(req.AttemptID) == "" || strings.TrimSpace(req.EventType) == "" {
		writeError(w, http.StatusBadRequest, "attempt_id and event_type are required")
		return
	}

	severity := strings.ToLower(strings.TrimSpace(req.Severity))
	action := "log_only"
	notice := "Exam monitoring event recorded."
	if severity == "critical" || severity == "high" {
		action = "pause_and_review"
		notice = "A monitoring issue was detected. Please wait for review or follow the displayed instruction."
	} else if severity == "medium" || severity == "warning" {
		action = "flag_for_review"
		notice = "A monitoring warning was recorded. Continue only if the issue is corrected."
	}

	writeJSON(w, http.StatusCreated, dto.LiveProctoringEventResponse{
		EventID:       fmt.Sprintf("live_event_%d", time.Now().UTC().UnixNano()),
		Accepted:      true,
		Action:        action,
		StudentNotice: notice,
	})
}

func uploadedEvidenceFiles(form *multipart.Form) map[string]bool {
	uploaded := map[string]bool{}
	if form == nil {
		return uploaded
	}

	for field, files := range form.File {
		for _, file := range files {
			if file.Filename == "" {
				continue
			}
			uploaded[field] = true
			uploaded[file.Filename] = true
		}
	}
	return uploaded
}

func fallbackPreExamReview(req dto.PreExamReviewRequest, uploaded map[string]bool) dto.PreExamReviewResponse {
	missing := []string{}
	lowLight := []string{}
	missingImages := []string{}
	riskLabels := []string{}
	audioRescan := false
	audioReview := false
	audioClipMissing := false

	for _, target := range req.Targets {
		if !target.Captured {
			missing = append(missing, target.Name)
		}
		if target.LightingScore < 0.08 {
			lowLight = append(lowLight, target.Name)
		}
		if target.ImageKey == "" || !uploaded[target.ImageKey] {
			missingImages = append(missingImages, target.Name)
		}
		for _, label := range target.Labels {
			if isReviewRiskLabel(label) {
				riskLabels = append(riskLabels, label)
			}
		}
	}

	findings := []dto.ReviewFindingResponse{}
	if len(missing) > 0 {
		findings = append(findings, dto.ReviewFindingResponse{
			Title:    "Coverage incomplete",
			Detail:   "Missing scan targets: " + strings.Join(uniqueStrings(missing), ", "),
			Severity: "warning",
		})
	} else {
		findings = append(findings, dto.ReviewFindingResponse{
			Title:    "360 coverage complete",
			Detail:   "All required areas were captured.",
			Severity: "success",
		})
	}

	if len(lowLight) > 0 {
		findings = append(findings, dto.ReviewFindingResponse{
			Title:    "Lighting concern",
			Detail:   "Low lighting detected in: " + strings.Join(uniqueStrings(lowLight), ", "),
			Severity: "warning",
		})
	} else {
		findings = append(findings, dto.ReviewFindingResponse{
			Title:    "Lighting accepted",
			Detail:   "Lighting is acceptable for review.",
			Severity: "success",
		})
	}

	if len(missingImages) > 0 {
		findings = append(findings, dto.ReviewFindingResponse{
			Title:    "Evidence record incomplete",
			Detail:   "Missing image uploads: " + strings.Join(uniqueStrings(missingImages), ", "),
			Severity: "warning",
		})
	}

	if len(riskLabels) > 0 {
		findings = append(findings, dto.ReviewFindingResponse{
			Title:    "Possible unauthorized item",
			Detail:   strings.Join(uniqueStrings(riskLabels), ", "),
			Severity: "warning",
		})
	}

	if req.Audio == nil {
		audioRescan = true
		findings = append(findings, dto.ReviewFindingResponse{
			Title:    "Sound check required",
			Detail:   "Room sound was not checked. Please repeat the sound check before exam startup.",
			Severity: "warning",
		})
	} else if !req.Audio.MicrophoneAvailable || !req.Audio.PermissionGranted || !req.Audio.InputLevelOK {
		audioRescan = true
		findings = append(findings, dto.ReviewFindingResponse{
			Title:    "Sound check required",
			Detail:   "Microphone access or input level needs correction.",
			Severity: "warning",
		})
	} else if !uploaded["audio_clip"] {
		audioClipMissing = true
		findings = append(findings, dto.ReviewFindingResponse{
			Title:    "Sound evidence incomplete",
			Detail:   "The microphone check passed but the audio clip was not uploaded.",
			Severity: "warning",
		})
	} else if audioRequiresReview(req.Audio) {
		audioReview = true
		findings = append(findings, dto.ReviewFindingResponse{
			Title:    "Room sound review required",
			Detail:   "Human voice, phone, notification, TV, radio, or conversation sound was detected.",
			Severity: "warning",
		})
	} else if !audioIsClassifiedOrAllowed(req.Audio) {
		audioRescan = true
		findings = append(findings, dto.ReviewFindingResponse{
			Title:    "Sound check unclear",
			Detail:   "Room sound could not be classified clearly. Please repeat the sound check.",
			Severity: "warning",
		})
	} else {
		findings = append(findings, dto.ReviewFindingResponse{
			Title:    "Sound check accepted",
			Detail:   "Room sound is acceptable for exam startup.",
			Severity: "success",
		})
	}

	decision := "approved"
	riskScore := 12
	riskLevel := "low"
	summary := "Pre-exam security check completed successfully."

	if len(missing) > 0 || len(lowLight) > 0 || len(missingImages) > 0 || audioRescan || audioClipMissing {
		decision = "rescan_required"
		riskScore = 45
		riskLevel = "medium"
		summary = "Some checks need correction before the exam can start."
	} else if len(riskLabels) > 0 || audioReview {
		decision = "review_required"
		riskScore = 62
		riskLevel = "medium"
		summary = "Review required before exam startup."
	}

	return dto.PreExamReviewResponse{
		ReviewID:  fmt.Sprintf("review_%d", time.Now().UTC().UnixNano()),
		Decision:  decision,
		RiskLevel: riskLevel,
		RiskScore: riskScore,
		Summary:   summary,
		Findings:  findings,
	}
}

func isReviewRiskLabel(label string) bool {
	value := strings.ToLower(label)
	riskTerms := []string{
		"phone",
		"book",
		"paper",
		"tablet",
		"screen",
		"earpiece",
		"headphone",
		"person",
	}

	for _, term := range riskTerms {
		if strings.Contains(value, term) {
			return true
		}
	}
	return false
}

func audioRequiresReview(audio *dto.PreExamReviewAudioRequest) bool {
	if audio == nil {
		return false
	}
	if audio.HumanVoiceDetected ||
		audio.PhoneRingDetected ||
		audio.NotificationDetected ||
		audio.TVOrRadioVoiceDetected {
		return true
	}

	label := strings.ToLower(audio.EnvironmentLabel)
	noiseClass := strings.ToLower(audio.DominantNoiseClass)
	reviewTerms := []string{
		"human_voice",
		"voice",
		"conversation",
		"phone_ring",
		"ringtone",
		"notification",
		"tv_voice",
		"radio_voice",
	}
	for _, term := range reviewTerms {
		if strings.Contains(label, term) || strings.Contains(noiseClass, term) {
			return true
		}
	}
	return audio.VoiceConfidence >= 0.70
}

func audioIsClassifiedOrAllowed(audio *dto.PreExamReviewAudioRequest) bool {
	if audio == nil {
		return false
	}
	if audio.AmbientNoiseAllowed {
		return true
	}

	noiseClass := strings.ToLower(strings.TrimSpace(audio.DominantNoiseClass))
	if noiseClass == "" || noiseClass == "unclassified" || noiseClass == "unknown" {
		return false
	}

	allowedClasses := []string{
		"fan",
		"generator",
		"rain",
		"traffic",
		"ac",
		"air_conditioner",
		"wind",
		"allowed_ambient_noise",
		"quiet_room",
		"quiet_environment",
		"moderate_environment",
		"noisy_environment",
		"very_noisy_environment",
	}
	for _, allowed := range allowedClasses {
		if strings.Contains(noiseClass, allowed) {
			return true
		}
	}
	return false
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}
