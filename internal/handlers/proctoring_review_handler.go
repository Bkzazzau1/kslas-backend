package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"kslasbackend/internal/dto"
	"kslasbackend/internal/services"
)

type ProctoringReviewHandler struct {
	openAIReview *services.OpenAIPreExamReviewService
}

func NewProctoringReviewHandler(openAIReview *services.OpenAIPreExamReviewService) *ProctoringReviewHandler {
	return &ProctoringReviewHandler{openAIReview: openAIReview}
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

	evidence := uploadedReviewEvidence(r.MultipartForm)
	fallback := fallbackPreExamReview(req, uploadedEvidenceFiles(evidence))
	if h.openAIReview != nil && h.openAIReview.Enabled() {
		review, err := h.openAIReview.ReviewPreExam(r.Context(), req, evidence, fallback)
		if err == nil {
			writeJSON(w, http.StatusOK, review)
			return
		}
		fallback.Source = "fallback"
		fallback.Issues = append(fallback.Issues, "AI review service unavailable: "+err.Error())
		fallback.Actions = append(fallback.Actions, "Please contact an invigilator if this message persists.")
		if fallback.Decision == "approved" {
			fallback.Decision = "review_required"
			fallback.RiskLevel = "medium"
			fallback.RiskScore = 55
			fallback.Summary = "Invigilator review required because the AI review service is unavailable."
		}
		writeJSON(w, http.StatusOK, fallback)
		return
	}

	fallback.Source = "fallback"
	writeJSON(w, http.StatusOK, fallback)
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

func uploadedReviewEvidence(form *multipart.Form) services.ProctoringReviewEvidence {
	evidence := services.ProctoringReviewEvidence{Files: []services.ProctoringEvidenceFile{}}
	if form == nil {
		return evidence
	}

	for field, files := range form.File {
		for _, header := range files {
			if header.Filename == "" {
				continue
			}
			file, err := header.Open()
			if err != nil {
				continue
			}
			data, _ := io.ReadAll(io.LimitReader(file, 10<<20))
			_ = file.Close()
			evidence.Files = append(evidence.Files, services.ProctoringEvidenceFile{
				FieldName:   field,
				FileName:    header.Filename,
				ContentType: header.Header.Get("Content-Type"),
				SizeBytes:   header.Size,
				Data:        data,
			})
		}
	}
	return evidence
}

func uploadedEvidenceFiles(evidence services.ProctoringReviewEvidence) map[string]bool {
	uploaded := map[string]bool{}
	for _, file := range evidence.Files {
		if file.FileName == "" {
			continue
		}
		uploaded[file.FieldName] = true
		uploaded[file.FileName] = true
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
	summary := "Backend review approved. Click OK to start the exam."

	if len(missing) > 0 || len(lowLight) > 0 || len(missingImages) > 0 || audioRescan || audioClipMissing {
		decision = "rescan_required"
		riskScore = 45
		riskLevel = "medium"
		summary = correctionSummary(findings)
	} else if len(riskLabels) > 0 || audioReview {
		decision = "review_required"
		riskScore = 62
		riskLevel = "medium"
		summary = reviewSummary(findings)
	}

	return dto.PreExamReviewResponse{
		ReviewID:  fmt.Sprintf("review_%d", time.Now().UTC().UnixNano()),
		Decision:  decision,
		RiskLevel: riskLevel,
		RiskScore: riskScore,
		Summary:   summary,
		Issues:    issueList(findings),
		Actions:   actionList(decision, findings),
		Source:    "fallback",
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

func correctionSummary(findings []dto.ReviewFindingResponse) string {
	issues := issueList(findings)
	if len(issues) == 0 {
		return "Correction required before the exam can start."
	}
	return "Correction required: " + strings.Join(issues, " ")
}

func reviewSummary(findings []dto.ReviewFindingResponse) string {
	issues := issueList(findings)
	if len(issues) == 0 {
		return "Invigilator review required before exam startup."
	}
	return "Invigilator review required: " + strings.Join(issues, " ")
}

func issueList(findings []dto.ReviewFindingResponse) []string {
	issues := []string{}
	for _, finding := range findings {
		severity := strings.ToLower(strings.TrimSpace(finding.Severity))
		if severity != "warning" && severity != "error" && severity != "critical" {
			continue
		}
		detail := strings.TrimSpace(finding.Detail)
		if detail == "" {
			detail = strings.TrimSpace(finding.Title)
		}
		if detail != "" {
			issues = append(issues, detail)
		}
	}
	return uniqueStrings(issues)
}

func actionList(decision string, findings []dto.ReviewFindingResponse) []string {
	if decision == "approved" {
		return []string{"Click OK to start the exam."}
	}

	actions := []string{}
	for _, issue := range issueList(findings) {
		lower := strings.ToLower(issue)
		switch {
		case strings.Contains(lower, "missing scan targets"):
			actions = append(actions, "Capture all required room scan targets again.")
		case strings.Contains(lower, "low lighting"):
			actions = append(actions, "Move to a brighter location or turn on more light.")
		case strings.Contains(lower, "missing image"):
			actions = append(actions, "Rescan so the missing image evidence can be uploaded.")
		case strings.Contains(lower, "microphone") || strings.Contains(lower, "sound") || strings.Contains(lower, "audio"):
			actions = append(actions, "Repeat the sound check in a quiet room with microphone access enabled.")
		case strings.Contains(lower, "unauthorized") || strings.Contains(lower, "phone") || strings.Contains(lower, "book"):
			actions = append(actions, "Remove unauthorized materials and request review again.")
		}
	}
	if len(actions) == 0 {
		actions = append(actions, "Correct the listed issue and run the security review again.")
	}
	return uniqueStrings(actions)
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
