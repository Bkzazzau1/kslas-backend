package dto

type PreExamReviewTargetRequest struct {
	Name          string   `json:"name"`
	Captured      bool     `json:"captured"`
	ImageKey      string   `json:"image_key"`
	LightingScore float64  `json:"lighting_score"`
	MotionScore   float64  `json:"motion_score"`
	SceneScore    float64  `json:"scene_score"`
	Labels        []string `json:"labels"`
}

type PreExamReviewRequest struct {
	StudentID  string                       `json:"student_id"`
	ExamID     string                       `json:"exam_id"`
	AttemptID  string                       `json:"attempt_id"`
	CapturedAt string                       `json:"captured_at"`
	Targets    []PreExamReviewTargetRequest `json:"targets"`
}

type ReviewFindingResponse struct {
	Title    string `json:"title"`
	Detail   string `json:"detail"`
	Severity string `json:"severity"`
}

type PreExamReviewResponse struct {
	ReviewID  string                  `json:"review_id"`
	Decision  string                  `json:"decision"`
	RiskLevel string                  `json:"risk_level"`
	RiskScore int                     `json:"risk_score"`
	Summary   string                  `json:"summary"`
	Findings  []ReviewFindingResponse `json:"findings"`
}
