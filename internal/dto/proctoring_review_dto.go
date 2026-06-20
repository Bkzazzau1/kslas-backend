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
	StudentID    string                       `json:"student_id"`
	ExamID       string                       `json:"exam_id"`
	AttemptID    string                       `json:"attempt_id"`
	CapturedAt   string                       `json:"captured_at"`
	FaceImageKey string                       `json:"face_image_key,omitempty"`
	SystemReview map[string]interface{}       `json:"system_review,omitempty"`
	FaceIdentity map[string]interface{}       `json:"face_identity,omitempty"`
	Audio        *PreExamReviewAudioRequest   `json:"audio,omitempty"`
	Targets      []PreExamReviewTargetRequest `json:"targets"`
}

type PreExamReviewAudioRequest struct {
	MicrophoneAvailable    bool    `json:"microphone_available"`
	PermissionGranted      bool    `json:"permission_granted"`
	InputLevelOK           bool    `json:"input_level_ok"`
	AverageRMS             float64 `json:"average_rms"`
	PeakRMS                float64 `json:"peak_rms"`
	VoiceConfidence        float64 `json:"voice_confidence"`
	EnvironmentLabel       string  `json:"environment_label"`
	DominantNoiseClass     string  `json:"dominant_noise_class,omitempty"`
	HumanVoiceDetected     bool    `json:"human_voice_detected"`
	PhoneRingDetected      bool    `json:"phone_ring_detected"`
	NotificationDetected   bool    `json:"notification_detected"`
	TVOrRadioVoiceDetected bool    `json:"tv_or_radio_voice_detected"`
	AmbientNoiseAllowed    bool    `json:"ambient_noise_allowed"`
	Message                string  `json:"message,omitempty"`
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
	Issues    []string                `json:"issues,omitempty"`
	Actions   []string                `json:"actions,omitempty"`
	Source    string                  `json:"source,omitempty"`
	Findings  []ReviewFindingResponse `json:"findings"`
}

type ExamStartApprovalRequest struct {
	StudentID     string                 `json:"student_id"`
	ExamID        string                 `json:"exam_id"`
	AttemptID     string                 `json:"attempt_id"`
	ManifestPath  string                 `json:"manifest_path"`
	FaceIDReady   bool                   `json:"face_id_ready"`
	RoomScanReady bool                   `json:"room_scan_ready"`
	AudioReady    bool                   `json:"audio_ready"`
	SystemReady   bool                   `json:"system_ready"`
	AudioReview   map[string]interface{} `json:"audio_review,omitempty"`
	SystemReview  map[string]interface{} `json:"system_review,omitempty"`
	Source        string                 `json:"source,omitempty"`
}

type ExamStartApprovalResponse struct {
	Status              string   `json:"status"`
	ApprovalSource      string   `json:"approval_source"`
	AIRecommendation    string   `json:"ai_recommendation"`
	RequiresHumanReview bool     `json:"requires_human_review"`
	ExamStartToken      string   `json:"exam_start_token,omitempty"`
	Message             string   `json:"message"`
	Issues              []string `json:"issues,omitempty"`
	ExpiresAt           string   `json:"expires_at,omitempty"`
}

type LiveProctoringEventRequest struct {
	StudentID string                 `json:"student_id"`
	ExamID    string                 `json:"exam_id"`
	AttemptID string                 `json:"attempt_id"`
	EventType string                 `json:"event_type"`
	Severity  string                 `json:"severity"`
	Message   string                 `json:"message"`
	CreatedAt string                 `json:"created_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type LiveProctoringEventResponse struct {
	EventID       string `json:"event_id"`
	Accepted      bool   `json:"accepted"`
	Action        string `json:"action"`
	StudentNotice string `json:"student_notice"`
}
