package dto

type RandomVideoSampleRequest struct {
	StudentID       string `json:"student_id"`
	ExamID          string `json:"exam_id"`
	AttemptID       string `json:"attempt_id"`
	SampleNumber    int    `json:"sample_number"`
	TotalSamples    int    `json:"total_samples"`
	DurationSeconds int    `json:"duration_seconds"`
	CapturedAt      string `json:"captured_at"`
	Purpose         string `json:"purpose"`
	ReviewTiming    string `json:"review_timing"`
}

type RandomVideoSampleResponse struct {
	SampleID        string `json:"sample_id"`
	StudentID       string `json:"student_id"`
	ExamID          string `json:"exam_id"`
	AttemptID       string `json:"attempt_id"`
	SampleNumber    int    `json:"sample_number"`
	TotalSamples    int    `json:"total_samples"`
	DurationSeconds int    `json:"duration_seconds"`
	CapturedAt      string `json:"captured_at"`
	StoredAt        string `json:"stored_at"`
	Purpose         string `json:"purpose"`
	ReviewTiming    string `json:"review_timing"`
	FileName        string `json:"file_name"`
	StoredPath      string `json:"stored_path"`
	PlaybackURL     string `json:"playback_url"`
	Status          string `json:"status"`
	Message         string `json:"message"`
}

type RandomVideoSampleListResponse struct {
	Items []RandomVideoSampleResponse `json:"items"`
	Count int                         `json:"count"`
}
