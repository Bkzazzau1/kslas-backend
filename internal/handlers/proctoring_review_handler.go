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

func uploadedEvidenceFiles(form *multipart.Form) map[string]bool {
	uploaded := map[string]bool{}
	if form == nil {
		return uploaded
	}

	for _, files := range form.File {
		for _, file := range files {
			if file.Filename == "" {
				continue
			}
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

	decision := "approved"
	riskScore := 12
	riskLevel := "low"
	summary := "Pre-exam security check completed successfully."

	if len(missing) > 0 || len(lowLight) > 0 || len(missingImages) > 0 {
		decision = "rescan_required"
		riskScore = 45
		riskLevel = "medium"
		summary = "Some checks need correction before the exam can start."
	} else if len(riskLabels) > 0 {
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
