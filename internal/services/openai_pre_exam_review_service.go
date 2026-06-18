package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"kslasbackend/internal/dto"
)

type ProctoringEvidenceFile struct {
	FieldName   string
	FileName    string
	ContentType string
	SizeBytes   int64
	Data        []byte
}

type ProctoringReviewEvidence struct {
	Files []ProctoringEvidenceFile
}

type OpenAIPreExamReviewService struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

func NewOpenAIPreExamReviewService(apiKey, baseURL, model string) *OpenAIPreExamReviewService {
	return &OpenAIPreExamReviewService{
		apiKey:  strings.TrimSpace(apiKey),
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		model:   strings.TrimSpace(model),
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

func (s *OpenAIPreExamReviewService) Enabled() bool {
	return s != nil && s.apiKey != "" && s.baseURL != "" && s.model != ""
}

func (s *OpenAIPreExamReviewService) ReviewPreExam(
	ctx context.Context,
	req dto.PreExamReviewRequest,
	evidence ProctoringReviewEvidence,
	fallback dto.PreExamReviewResponse,
) (dto.PreExamReviewResponse, error) {
	if !s.Enabled() {
		return fallback, errors.New("openai review is not configured")
	}

	payload := map[string]interface{}{
		"model": s.model,
		"input": []map[string]interface{}{
			{
				"role": "system",
				"content": []map[string]interface{}{
					{
						"type": "input_text",
						"text": "You are a strict but fair university pre-exam proctoring reviewer. Review the provided scan images, face identity metadata, audio metadata/evidence summary, and system review. Return only JSON matching the schema. Allow the exam only when the candidate identity is plausible, required room coverage is complete, evidence is present, audio is acceptable, and there are no unauthorized items or people. If not allowed, state clear student-facing issues and actions.",
					},
				},
			},
			{
				"role":    "user",
				"content": s.reviewContent(req, evidence, fallback),
			},
		},
		"text": map[string]interface{}{
			"format": map[string]interface{}{
				"type":   "json_schema",
				"name":   "kslas_pre_exam_review",
				"strict": true,
				"schema": map[string]interface{}{
					"type":                 "object",
					"additionalProperties": false,
					"required": []string{
						"decision",
						"risk_level",
						"risk_score",
						"summary",
						"issues",
						"actions",
						"findings",
					},
					"properties": map[string]interface{}{
						"decision": map[string]interface{}{
							"type": "string",
							"enum": []string{"approved", "rescan_required", "review_required"},
						},
						"risk_level": map[string]interface{}{
							"type": "string",
							"enum": []string{"low", "medium", "high"},
						},
						"risk_score": map[string]interface{}{
							"type":    "integer",
							"minimum": 0,
							"maximum": 100,
						},
						"summary": map[string]interface{}{"type": "string"},
						"issues": map[string]interface{}{
							"type":  "array",
							"items": map[string]interface{}{"type": "string"},
						},
						"actions": map[string]interface{}{
							"type":  "array",
							"items": map[string]interface{}{"type": "string"},
						},
						"findings": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"type":                 "object",
								"additionalProperties": false,
								"required":             []string{"title", "detail", "severity"},
								"properties": map[string]interface{}{
									"title":    map[string]interface{}{"type": "string"},
									"detail":   map[string]interface{}{"type": "string"},
									"severity": map[string]interface{}{"type": "string"},
								},
							},
						},
					},
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fallback, fmt.Errorf("marshal openai review request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/responses", bytes.NewReader(body))
	if err != nil {
		return fallback, fmt.Errorf("create openai review request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return fallback, fmt.Errorf("send openai review request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return fallback, fmt.Errorf("read openai review response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fallback, fmt.Errorf("openai review failed: %s", strings.TrimSpace(string(respBody)))
	}

	review, err := parseOpenAIReviewResponse(respBody)
	if err != nil {
		return fallback, err
	}

	review.ReviewID = fallback.ReviewID
	review.Source = "openai"
	return review, nil
}

func (s *OpenAIPreExamReviewService) reviewContent(
	req dto.PreExamReviewRequest,
	evidence ProctoringReviewEvidence,
	fallback dto.PreExamReviewResponse,
) []map[string]interface{} {
	content := []map[string]interface{}{
		{
			"type": "input_text",
			"text": reviewPromptText(req, evidence, fallback),
		},
	}

	for _, file := range evidence.Files {
		if !isImageEvidence(file) || len(file.Data) == 0 {
			continue
		}
		content = append(content, map[string]interface{}{
			"type":      "input_image",
			"image_url": imageDataURL(file),
		})
	}

	return content
}

func reviewPromptText(
	req dto.PreExamReviewRequest,
	evidence ProctoringReviewEvidence,
	fallback dto.PreExamReviewResponse,
) string {
	reviewPackage := map[string]interface{}{
		"student_id":        req.StudentID,
		"exam_id":           req.ExamID,
		"attempt_id":        req.AttemptID,
		"captured_at":       req.CapturedAt,
		"face_image_key":    req.FaceImageKey,
		"face_identity":     req.FaceIdentity,
		"system_review":     req.SystemReview,
		"audio":             req.Audio,
		"targets":           req.Targets,
		"evidence_files":    evidenceFileSummary(evidence.Files),
		"local_rule_result": fallback,
	}
	data, _ := json.MarshalIndent(reviewPackage, "", "  ")
	return "Review this K-SLAS pre-exam evidence package. The attached images are the uploaded room/face evidence. Audio is represented by the uploaded audio clip metadata and audio metrics in JSON. Return approved only when safe. If not approved, explain exactly what the student must correct.\n\n" + string(data)
}

func evidenceFileSummary(files []ProctoringEvidenceFile) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(files))
	for _, file := range files {
		out = append(out, map[string]interface{}{
			"field_name":        file.FieldName,
			"file_name":         file.FileName,
			"content_type":      file.ContentType,
			"size_bytes":        file.SizeBytes,
			"included_as_image": isImageEvidence(file),
			"is_audio_clip":     file.FieldName == "audio_clip",
		})
	}
	return out
}

func parseOpenAIReviewResponse(data []byte) (dto.PreExamReviewResponse, error) {
	var raw struct {
		OutputText string `json:"output_text"`
		Output     []struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"output"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return dto.PreExamReviewResponse{}, fmt.Errorf("decode openai review response: %w", err)
	}

	text := strings.TrimSpace(raw.OutputText)
	if text == "" {
		for _, output := range raw.Output {
			for _, content := range output.Content {
				if content.Text != "" {
					text = strings.TrimSpace(content.Text)
					break
				}
			}
			if text != "" {
				break
			}
		}
	}
	if text == "" {
		return dto.PreExamReviewResponse{}, errors.New("openai review returned no text")
	}

	var review dto.PreExamReviewResponse
	if err := json.Unmarshal([]byte(text), &review); err != nil {
		return dto.PreExamReviewResponse{}, fmt.Errorf("decode openai review json: %w", err)
	}
	review.Decision = normalizeReviewDecision(review.Decision)
	if review.RiskLevel == "" {
		review.RiskLevel = "medium"
	}
	if review.Summary == "" {
		review.Summary = "Review completed."
	}
	return review, nil
}

func normalizeReviewDecision(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "approved":
		return "approved"
	case "rescan_required":
		return "rescan_required"
	default:
		return "review_required"
	}
}

func isImageEvidence(file ProctoringEvidenceFile) bool {
	contentType := strings.ToLower(file.ContentType)
	if strings.HasPrefix(contentType, "image/") {
		return true
	}
	ext := strings.ToLower(filepath.Ext(file.FileName))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp"
}

func imageDataURL(file ProctoringEvidenceFile) string {
	contentType := strings.TrimSpace(file.ContentType)
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(file.FileName))
	}
	if contentType == "" {
		contentType = "image/jpeg"
	}
	return "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(file.Data)
}
