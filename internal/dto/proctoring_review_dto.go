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
	Audio      *PreExamReviewAudioRequest   `json:"audio,omitempty"`
	Targets    []PreExamReviewTargetRequest `json:"targets"`
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
	Findings  []ReviewFindingResponse `json:"findings"`
}
