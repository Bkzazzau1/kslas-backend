package dto

import "time"

type InvigilatorEvidenceMetricsResponse struct {
	OpenEvidence   int `json:"open_evidence"`
	Captured       int `json:"captured"`
	PendingCapture int `json:"pending_capture"`
	Critical       int `json:"critical"`
	DraftReports   int `json:"draft_reports"`
}

type InvigilatorEvidenceCaseResponse struct {
	ID             string    `json:"id"`
	Candidate      string    `json:"candidate"`
	Matric         string    `json:"matric"`
	Course         string    `json:"course"`
	Session        string    `json:"session"`
	EventType      string    `json:"event_type"`
	Severity       string    `json:"severity"`
	Status         string    `json:"status"`
	RiskScore      int       `json:"risk_score"`
	Confidence     float64   `json:"confidence"`
	EvidenceTypes  []string  `json:"evidence_types"`
	EvidenceStatus string    `json:"evidence_status"`
	EvidencePath   string    `json:"evidence_path"`
	Time           string    `json:"time"`
	Recommendation string    `json:"recommendation"`
	Decision       string    `json:"decision"`
	CreatedAt      time.Time `json:"created_at"`
}

type InvigilatorEvidenceQueueResponse struct {
	Metrics InvigilatorEvidenceMetricsResponse `json:"metrics"`
	Items   []InvigilatorEvidenceCaseResponse `json:"items"`
	Count   int                               `json:"count"`
}

type InvigilatorEvidenceDecisionRequest struct {
	Action string `json:"action"`
	Note   string `json:"note"`
}

type InvigilatorEvidenceDecisionResponse struct {
	CaseID    string    `json:"case_id"`
	Action    string    `json:"action"`
	Note      string    `json:"note,omitempty"`
	Status    string    `json:"status"`
	Decision  string    `json:"decision"`
	DecidedAt time.Time `json:"decided_at"`
}
