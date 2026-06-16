package handlers

import (
	"net/http"
	"strings"
	"time"

	"kslasbackend/internal/dto"
)

type InvigilatorEvidenceHandler struct{}

func NewInvigilatorEvidenceHandler() *InvigilatorEvidenceHandler {
	return &InvigilatorEvidenceHandler{}
}

func (h *InvigilatorEvidenceHandler) Queue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	severity := strings.TrimSpace(r.URL.Query().Get("severity"))
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	evidenceType := strings.TrimSpace(r.URL.Query().Get("evidence_type"))
	items := filterEvidenceCases(mockEvidenceCases(), severity, status, evidenceType)

	writeJSON(w, http.StatusOK, dto.InvigilatorEvidenceQueueResponse{
		Metrics: mockEvidenceMetrics(),
		Items:   items,
		Count:   len(items),
	})
}

func (h *InvigilatorEvidenceHandler) Decision(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

	caseID := strings.TrimSpace(r.PathValue("caseID"))
	if caseID == "" {
		writeError(w, http.StatusBadRequest, "case id is required")
		return
	}

	var req dto.InvigilatorEvidenceDecisionRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	action := strings.ToLower(strings.TrimSpace(req.Action))
	if action == "" {
		writeError(w, http.StatusBadRequest, "action is required")
		return
	}

	status, decision := decisionStatus(action)
	if status == "" {
		writeError(w, http.StatusBadRequest, "unsupported decision action")
		return
	}

	writeJSON(w, http.StatusOK, dto.InvigilatorEvidenceDecisionResponse{
		CaseID:    caseID,
		Action:    action,
		Note:      strings.TrimSpace(req.Note),
		Status:    status,
		Decision:  decision,
		DecidedAt: time.Now().UTC(),
	})
}

func filterEvidenceCases(items []dto.InvigilatorEvidenceCaseResponse, severity, status, evidenceType string) []dto.InvigilatorEvidenceCaseResponse {
	filtered := make([]dto.InvigilatorEvidenceCaseResponse, 0, len(items))
	for _, item := range items {
		if severity != "" && !strings.EqualFold(severity, "all") && !strings.EqualFold(item.Severity, severity) {
			continue
		}
		if status != "" && !strings.EqualFold(status, "all") && !strings.EqualFold(item.Status, status) {
			continue
		}
		if evidenceType != "" && !strings.EqualFold(evidenceType, "all") && !containsFold(item.EvidenceTypes, evidenceType) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func containsFold(items []string, value string) bool {
	for _, item := range items {
		if strings.EqualFold(item, value) {
			return true
		}
	}
	return false
}

func decisionStatus(action string) (string, string) {
	switch action {
	case "clear", "clear_candidate":
		return "Cleared", "Candidate cleared after evidence review"
	case "warn", "issue_warning":
		return "Warning issued", "Warning issued by invigilator"
	case "escalate", "escalate_report":
		return "Escalated", "Escalated to exam officer"
	case "malpractice", "draft_malpractice":
		return "Malpractice draft", "Draft malpractice report opened"
	case "close", "close_review":
		return "Closed", "Evidence review closed"
	default:
		return "", ""
	}
}

func mockEvidenceMetrics() dto.InvigilatorEvidenceMetricsResponse {
	return dto.InvigilatorEvidenceMetricsResponse{
		OpenEvidence:   24,
		Captured:       18,
		PendingCapture: 4,
		Critical:       3,
		DraftReports:   5,
	}
}

func mockEvidenceCases() []dto.InvigilatorEvidenceCaseResponse {
	now := time.Now().UTC()
	return []dto.InvigilatorEvidenceCaseResponse{
		{
			ID:             "case-001",
			Candidate:      "Aisha Musa",
			Matric:         "KASU/CSC/021",
			Course:         "CSC 309 Artificial Intelligence",
			Session:        "DLC Online Proctoring Group A",
			EventType:      "Multiple faces detected",
			Severity:       "High",
			Status:         "Pending review",
			RiskScore:      86,
			Confidence:     0.94,
			EvidenceTypes:  []string{"Camera", "Manifest"},
			EvidenceStatus: "Captured",
			EvidencePath:   "evidence://KASU/CSC/021/session-309/case-001.json",
			Time:           "Today, 10:38",
			Recommendation: "Review camera frame evidence and escalate if second person remains visible.",
			Decision:       "No decision yet",
			CreatedAt:      now.Add(-25 * time.Minute),
		},
		{
			ID:             "case-002",
			Candidate:      "Bello Adamu",
			Matric:         "KASU/CSC/044",
			Course:         "CSC 309 Artificial Intelligence",
			Session:        "DLC Online Proctoring Group A",
			EventType:      "Human voice detected",
			Severity:       "High",
			Status:         "Escalated",
			RiskScore:      74,
			Confidence:     0.89,
			EvidenceTypes:  []string{"Audio", "Manifest"},
			EvidenceStatus: "Captured",
			EvidencePath:   "evidence://KASU/CSC/044/session-309/case-002.json",
			Time:           "Today, 10:42",
			Recommendation: "Listen to the short audio clip and compare with mouth movement timeline.",
			Decision:       "Escalated to exam officer",
			CreatedAt:      now.Add(-21 * time.Minute),
		},
		{
			ID:             "case-003",
			Candidate:      "Maryam Sani",
			Matric:         "KASU/CSC/078",
			Course:         "GST 211 Communication Skills",
			Session:        "Morning CBT Block",
			EventType:      "Tab switching detected",
			Severity:       "Medium",
			Status:         "Pending review",
			RiskScore:      46,
			Confidence:     0.78,
			EvidenceTypes:  []string{"Screenshot", "Manifest"},
			EvidenceStatus: "Pending capture",
			EvidencePath:   "evidence://KASU/CSC/078/session-gst/case-003.json",
			Time:           "Today, 10:51",
			Recommendation: "Confirm whether the system app lost focus or candidate attempted navigation.",
			Decision:       "No decision yet",
			CreatedAt:      now.Add(-12 * time.Minute),
		},
		{
			ID:             "case-004",
			Candidate:      "Usman Ibrahim",
			Matric:         "KASU/BUS/012",
			Course:         "ACC 201 Financial Accounting",
			Session:        "Hybrid Practical Session",
			EventType:      "Phone detected",
			Severity:       "Critical",
			Status:         "Malpractice draft",
			RiskScore:      112,
			Confidence:     0.97,
			EvidenceTypes:  []string{"Camera", "Screenshot", "Manifest"},
			EvidenceStatus: "Captured",
			EvidencePath:   "evidence://KASU/BUS/012/session-acc/case-004.json",
			Time:           "Today, 11:06",
			Recommendation: "Prepare malpractice report if physical invigilator confirms device presence.",
			Decision:       "Draft report opened",
			CreatedAt:      now.Add(-3 * time.Minute),
		},
	}
}
