package dto

type FaceEnrollmentImageRequest struct {
	Field        string  `json:"field"`
	PoseCode     string  `json:"pose_code"`
	Title        string  `json:"title"`
	Instruction  string  `json:"instruction"`
	QualityScore float64 `json:"quality_score"`
	FileName     string  `json:"file_name"`
}

type FaceEnrollmentRequest struct {
	StudentID               string                       `json:"student_id"`
	CapturedAt              string                       `json:"captured_at"`
	RequiredImages          int                          `json:"required_images"`
	Purpose                 string                       `json:"purpose"`
	ReviewableByInvigilator bool                         `json:"reviewable_by_invigilator"`
	Images                  []FaceEnrollmentImageRequest `json:"images"`
}

type FaceEnrollmentImageResponse struct {
	Field        string  `json:"field"`
	PoseCode     string  `json:"pose_code"`
	Title        string  `json:"title"`
	Instruction  string  `json:"instruction"`
	QualityScore float64 `json:"quality_score"`
	FileName     string  `json:"file_name"`
	StoredPath   string  `json:"stored_path"`
	ViewURL      string  `json:"view_url"`
}

type FaceEnrollmentResponse struct {
	EnrollmentID            string                        `json:"enrollment_id"`
	StudentID               string                        `json:"student_id"`
	Status                  string                        `json:"status"`
	Message                 string                        `json:"message"`
	CapturedAt              string                        `json:"captured_at"`
	StoredAt                string                        `json:"stored_at"`
	RequiredImages          int                           `json:"required_images"`
	UploadedImages          int                           `json:"uploaded_images"`
	Purpose                 string                        `json:"purpose"`
	ReviewableByInvigilator bool                          `json:"reviewable_by_invigilator"`
	ManifestPath            string                        `json:"manifest_path"`
	Images                  []FaceEnrollmentImageResponse `json:"images"`
}

type FaceEnrollmentListResponse struct {
	Items []FaceEnrollmentResponse `json:"items"`
	Count int                      `json:"count"`
}
