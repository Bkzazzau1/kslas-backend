package dto

import "time"

type CourseMaterialCreateRequest struct {
	CourseID      uint   `json:"course_id"`
	CourseCode    string `json:"course_code"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	MaterialType  string `json:"material_type"`
	ExternalURL   string `json:"external_url"`
	AllowDownload *bool  `json:"allow_download"`
	Publish       *bool  `json:"publish"`
}

type CourseMaterialUpdateRequest struct {
	Title         string `json:"title"`
	Description   string `json:"description"`
	MaterialType  string `json:"material_type"`
	ExternalURL   string `json:"external_url"`
	AllowDownload *bool  `json:"allow_download"`
	Publish       *bool  `json:"publish"`
}

type CourseMaterialResponse struct {
	ID               uint       `json:"id"`
	UUID             string     `json:"uuid"`
	CourseID         uint       `json:"course_id"`
	CourseCode       string     `json:"course_code,omitempty"`
	CourseTitle      string     `json:"course_title,omitempty"`
	Title            string     `json:"title"`
	Description      string     `json:"description,omitempty"`
	MaterialType     string     `json:"material_type"`
	OriginalFileName string     `json:"original_file_name,omitempty"`
	ExternalURL      string     `json:"external_url,omitempty"`
	MimeType         string     `json:"mime_type,omitempty"`
	SizeBytes        int64      `json:"size_bytes"`
	AllowDownload    bool       `json:"allow_download"`
	UploadedBy       uint       `json:"uploaded_by"`
	PublishedAt      *time.Time `json:"published_at,omitempty"`
	FileURL          string     `json:"file_url,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type VideoLectureCreateRequest struct {
	CourseID           uint     `json:"course_id"`
	CourseCode         string   `json:"course_code"`
	Title              string   `json:"title"`
	Subtitle           string   `json:"subtitle"`
	Description        string   `json:"description"`
	LecturerName       string   `json:"lecturer_name"`
	ExternalURL        string   `json:"external_url"`
	DurationMinutes    uint     `json:"duration_minutes"`
	AudienceKeys       []string `json:"audience_keys"`
	Tags               []string `json:"tags"`
	AllowDownload      *bool    `json:"allow_download"`
	RequireWatchedMark *bool    `json:"require_watched_mark"`
}

type VideoLectureResponse struct {
	ID                 uint      `json:"id"`
	UUID               string    `json:"uuid"`
	CourseID           uint      `json:"course_id"`
	CourseCode         string    `json:"course_code,omitempty"`
	CourseTitle        string    `json:"course_title,omitempty"`
	Title              string    `json:"title"`
	Subtitle           string    `json:"subtitle,omitempty"`
	Description        string    `json:"description,omitempty"`
	LecturerName       string    `json:"lecturer_name,omitempty"`
	SourceType         string    `json:"source_type"`
	OriginalFileName   string    `json:"original_file_name,omitempty"`
	ExternalURL        string    `json:"external_url,omitempty"`
	MimeType           string    `json:"mime_type,omitempty"`
	SizeBytes          int64     `json:"size_bytes"`
	DurationMinutes    uint      `json:"duration_minutes"`
	AudienceKeys       []string  `json:"audience_keys,omitempty"`
	Tags               []string  `json:"tags,omitempty"`
	AllowDownload      bool      `json:"allow_download"`
	RequireWatchedMark bool      `json:"require_watched_mark"`
	UploadedBy         uint      `json:"uploaded_by"`
	PublishedAt        time.Time `json:"published_at"`
	StreamURL          string    `json:"stream_url,omitempty"`
	WatchedCount       int       `json:"watched_count"`
	WatchedByUserIDs   []uint    `json:"watched_by_user_ids,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type VideoLectureWatchRequest struct {
	WatchedAt *time.Time `json:"watched_at"`
}

type LiveSessionSettingsPayload struct {
	StudentCameraRequired  bool `json:"student_camera_required"`
	CaptureRegistrationNo  bool `json:"capture_registration_number"`
	AllowStudentRecording  bool `json:"allow_student_recording"`
	AllowLecturerRecording bool `json:"allow_lecturer_recording"`
	AttendanceEnabled      bool `json:"attendance_enabled"`
	ChatEnabled            bool `json:"chat_enabled"`
	QuestionsEnabled       bool `json:"questions_enabled"`
}

type LiveSessionCreateRequest struct {
	CourseID     uint                       `json:"course_id"`
	Title        string                     `json:"title"`
	Description  string                     `json:"description"`
	LecturerName string                     `json:"lecturer_name"`
	RoomName     string                     `json:"room_name"`
	StartTime    time.Time                  `json:"start_time"`
	EndTime      time.Time                  `json:"end_time"`
	Status       string                     `json:"status"`
	Agenda       []string                   `json:"agenda"`
	Materials    []string                   `json:"materials"`
	Settings     LiveSessionSettingsPayload `json:"settings"`
}

type LiveSessionUpdateRequest struct {
	Title        string                     `json:"title"`
	Description  string                     `json:"description"`
	LecturerName string                     `json:"lecturer_name"`
	RoomName     string                     `json:"room_name"`
	StartTime    time.Time                  `json:"start_time"`
	EndTime      time.Time                  `json:"end_time"`
	Status       string                     `json:"status"`
	Agenda       []string                   `json:"agenda"`
	Materials    []string                   `json:"materials"`
	Settings     LiveSessionSettingsPayload `json:"settings"`
}

type LiveSessionResponse struct {
	ID           uint                       `json:"id"`
	UUID         string                     `json:"uuid"`
	CourseID     uint                       `json:"course_id"`
	CourseCode   string                     `json:"course_code,omitempty"`
	CourseTitle  string                     `json:"course_title,omitempty"`
	Title        string                     `json:"title"`
	Description  string                     `json:"description,omitempty"`
	LecturerName string                     `json:"lecturer_name,omitempty"`
	RoomName     string                     `json:"room_name"`
	StartTime    time.Time                  `json:"start_time"`
	EndTime      time.Time                  `json:"end_time"`
	Status       string                     `json:"status"`
	Agenda       []string                   `json:"agenda,omitempty"`
	Materials    []string                   `json:"materials,omitempty"`
	Settings     LiveSessionSettingsPayload `json:"settings"`
	CreatedBy    uint                       `json:"created_by"`
	CreatedAt    time.Time                  `json:"created_at"`
	UpdatedAt    time.Time                  `json:"updated_at"`
}

type AttendanceUpsertRequest struct {
	UserID             uint       `json:"user_id"`
	RegistrationNumber string     `json:"registration_number"`
	Status             string     `json:"status"`
	JoinedAt           *time.Time `json:"joined_at"`
	LeftAt             *time.Time `json:"left_at"`
	DurationMinutes    uint       `json:"duration_minutes"`
}

type AttendanceBulkUpsertRequest struct {
	Items []AttendanceUpsertRequest `json:"items"`
}

type AttendanceResponse struct {
	ID                 uint       `json:"id"`
	LiveSessionID      uint       `json:"live_session_id"`
	UserID             uint       `json:"user_id"`
	RegistrationNumber string     `json:"registration_number,omitempty"`
	Status             string     `json:"status"`
	JoinedAt           *time.Time `json:"joined_at,omitempty"`
	LeftAt             *time.Time `json:"left_at,omitempty"`
	DurationMinutes    uint       `json:"duration_minutes"`
	CapturedBy         uint       `json:"captured_by"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type LiveSessionRecordingCreateRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ExternalURL string `json:"external_url"`
}

type LiveSessionRecordingResponse struct {
	ID               uint      `json:"id"`
	UUID             string    `json:"uuid"`
	LiveSessionID    uint      `json:"live_session_id"`
	Title            string    `json:"title"`
	Description      string    `json:"description,omitempty"`
	SourceType       string    `json:"source_type"`
	OriginalFileName string    `json:"original_file_name,omitempty"`
	ExternalURL      string    `json:"external_url,omitempty"`
	MimeType         string    `json:"mime_type,omitempty"`
	SizeBytes        int64     `json:"size_bytes"`
	AddedBy          uint      `json:"added_by"`
	PublishedAt      time.Time `json:"published_at"`
	FileURL          string    `json:"file_url,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CourseForumPostCreateRequest struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	ParentID *uint  `json:"parent_id"`
}

type CourseForumPostModerationRequest struct {
	IsPinned *bool `json:"is_pinned"`
	IsLocked *bool `json:"is_locked"`
}

type CourseForumPostResponse struct {
	ID                uint                      `json:"id"`
	UUID              string                    `json:"uuid"`
	CourseID          uint                      `json:"course_id"`
	CourseCode        string                    `json:"course_code,omitempty"`
	CourseTitle       string                    `json:"course_title,omitempty"`
	AuthorID          uint                      `json:"author_id"`
	AuthorDisplayName string                    `json:"author_display_name"`
	AuthorRole        string                    `json:"author_role"`
	Title             string                    `json:"title,omitempty"`
	Body              string                    `json:"body"`
	IsPinned          bool                      `json:"is_pinned"`
	IsLocked          bool                      `json:"is_locked"`
	ParentID          *uint                     `json:"parent_id,omitempty"`
	Replies           []CourseForumPostResponse `json:"replies,omitempty"`
	CreatedAt         time.Time                 `json:"created_at"`
	UpdatedAt         time.Time                 `json:"updated_at"`
}

type CourseDirectMessageCreateRequest struct {
	RecipientID uint   `json:"recipient_id"`
	Body        string `json:"body"`
}

type CourseDirectMessageResponse struct {
	ID                   uint       `json:"id"`
	UUID                 string     `json:"uuid"`
	CourseID             uint       `json:"course_id"`
	CourseCode           string     `json:"course_code,omitempty"`
	CourseTitle          string     `json:"course_title,omitempty"`
	SenderID             uint       `json:"sender_id"`
	SenderDisplayName    string     `json:"sender_display_name,omitempty"`
	RecipientID          uint       `json:"recipient_id"`
	RecipientDisplayName string     `json:"recipient_display_name,omitempty"`
	Body                 string     `json:"body"`
	ReadAt               *time.Time `json:"read_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

type AssignmentCreateRequest struct {
	CourseID           uint                              `json:"course_id"`
	CourseCode         string                            `json:"course_code"`
	Title              string                            `json:"title"`
	Description        string                            `json:"description"`
	Instructions       string                            `json:"instructions"`
	DueAt              *time.Time                        `json:"due_at"`
	MaxScore           float64                           `json:"max_score"`
	AssignmentType     string                            `json:"assignment_type"`
	SubmissionMode     string                            `json:"submission_mode"`
	AllowedExtensions  []string                          `json:"allowed_extensions"`
	WhiteboardEnabled  bool                              `json:"whiteboard_enabled"`
	WhiteboardRequired bool                              `json:"whiteboard_required"`
	WhiteboardPrompt   string                            `json:"whiteboard_prompt"`
	PeerReviewEnabled  bool                              `json:"peer_review_enabled"`
	PeerReviewRubric   []string                          `json:"peer_review_rubric"`
	GroupSource        string                            `json:"group_source"`
	Groups             []AssignmentGroupCreateInput      `json:"groups"`
	PeerReviews        []AssignmentPeerReviewAssignInput `json:"peer_reviews"`
	Status             string                            `json:"status"`
}

type AssignmentUpdateRequest struct {
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	Instructions       string     `json:"instructions"`
	DueAt              *time.Time `json:"due_at"`
	MaxScore           float64    `json:"max_score"`
	AssignmentType     string     `json:"assignment_type"`
	SubmissionMode     string     `json:"submission_mode"`
	AllowedExtensions  []string   `json:"allowed_extensions"`
	WhiteboardEnabled  bool       `json:"whiteboard_enabled"`
	WhiteboardRequired bool       `json:"whiteboard_required"`
	WhiteboardPrompt   string     `json:"whiteboard_prompt"`
	PeerReviewEnabled  bool       `json:"peer_review_enabled"`
	PeerReviewRubric   []string   `json:"peer_review_rubric"`
	GroupSource        string     `json:"group_source"`
	Status             string     `json:"status"`
}

type AssignmentResponse struct {
	ID                 uint                           `json:"id"`
	UUID               string                         `json:"uuid"`
	CourseID           uint                           `json:"course_id"`
	CourseCode         string                         `json:"course_code,omitempty"`
	CourseTitle        string                         `json:"course_title,omitempty"`
	Title              string                         `json:"title"`
	Description        string                         `json:"description,omitempty"`
	Instructions       string                         `json:"instructions,omitempty"`
	DueAt              *time.Time                     `json:"due_at,omitempty"`
	MaxScore           float64                        `json:"max_score"`
	AssignmentType     string                         `json:"assignment_type"`
	SubmissionMode     string                         `json:"submission_mode"`
	AllowedExtensions  []string                       `json:"allowed_extensions,omitempty"`
	WhiteboardEnabled  bool                           `json:"whiteboard_enabled"`
	WhiteboardRequired bool                           `json:"whiteboard_required"`
	WhiteboardPrompt   string                         `json:"whiteboard_prompt,omitempty"`
	PeerReviewEnabled  bool                           `json:"peer_review_enabled"`
	PeerReviewRubric   []string                       `json:"peer_review_rubric,omitempty"`
	GroupSource        string                         `json:"group_source,omitempty"`
	Groups             []AssignmentGroupResponse      `json:"groups,omitempty"`
	PeerReviews        []AssignmentPeerReviewResponse `json:"peer_reviews,omitempty"`
	SubmissionCount    int                            `json:"submission_count"`
	GradeCount         int                            `json:"grade_count"`
	Status             string                         `json:"status"`
	CreatedBy          uint                           `json:"created_by"`
	CreatedAt          time.Time                      `json:"created_at"`
	UpdatedAt          time.Time                      `json:"updated_at"`
}

type AssignmentGroupCreateInput struct {
	Name       string `json:"name"`
	Source     string `json:"source"`
	StudentIDs []uint `json:"student_ids"`
}

type AssignmentGroupResponse struct {
	ID           uint                            `json:"id"`
	UUID         string                          `json:"uuid"`
	AssignmentID uint                            `json:"assignment_id"`
	Name         string                          `json:"name"`
	Source       string                          `json:"source"`
	Members      []AssignmentGroupMemberResponse `json:"members,omitempty"`
	CreatedBy    uint                            `json:"created_by"`
	CreatedAt    time.Time                       `json:"created_at"`
	UpdatedAt    time.Time                       `json:"updated_at"`
}

type AssignmentGroupMemberResponse struct {
	ID        uint      `json:"id"`
	GroupID   uint      `json:"group_id"`
	StudentID uint      `json:"student_id"`
	AddedBy   uint      `json:"added_by"`
	CreatedAt time.Time `json:"created_at"`
}

type AssignmentSubmissionCreateRequest struct {
	TextAnswer     string                          `json:"text_answer"`
	GroupID        *uint                           `json:"group_id"`
	WhiteboardData string                          `json:"whiteboard_data"`
	Files          []AssignmentSubmissionFileInput `json:"files"`
}

type AssignmentSubmissionFileInput struct {
	StoragePath      string `json:"storage_path"`
	OriginalFileName string `json:"original_file_name"`
	StoredFileName   string `json:"stored_file_name"`
	MimeType         string `json:"mime_type"`
	SizeBytes        int64  `json:"size_bytes"`
}

type AssignmentSubmissionResponse struct {
	ID             uint                               `json:"id"`
	UUID           string                             `json:"uuid"`
	AssignmentID   uint                               `json:"assignment_id"`
	StudentID      uint                               `json:"student_id"`
	GroupID        *uint                              `json:"group_id,omitempty"`
	TextAnswer     string                             `json:"text_answer,omitempty"`
	WhiteboardData string                             `json:"whiteboard_data,omitempty"`
	Status         string                             `json:"status"`
	SubmittedAt    time.Time                          `json:"submitted_at"`
	Files          []AssignmentSubmissionFileResponse `json:"files,omitempty"`
	CreatedAt      time.Time                          `json:"created_at"`
	UpdatedAt      time.Time                          `json:"updated_at"`
}

type AssignmentSubmissionFileResponse struct {
	ID               uint      `json:"id"`
	SubmissionID     uint      `json:"submission_id"`
	StoragePath      string    `json:"storage_path"`
	OriginalFileName string    `json:"original_file_name"`
	StoredFileName   string    `json:"stored_file_name"`
	MimeType         string    `json:"mime_type,omitempty"`
	SizeBytes        int64     `json:"size_bytes"`
	CreatedAt        time.Time `json:"created_at"`
}

type AssignmentPeerReviewAssignInput struct {
	ReviewerID         uint  `json:"reviewer_id"`
	TargetStudentID    uint  `json:"target_student_id"`
	TargetSubmissionID *uint `json:"target_submission_id"`
}

type AssignmentPeerReviewSubmitRequest struct {
	Score        float64         `json:"score"`
	Feedback     string          `json:"feedback"`
	RubricChecks map[string]bool `json:"rubric_checks"`
}

type AssignmentPeerReviewResponse struct {
	ID                 uint            `json:"id"`
	UUID               string          `json:"uuid"`
	AssignmentID       uint            `json:"assignment_id"`
	ReviewerID         uint            `json:"reviewer_id"`
	TargetStudentID    uint            `json:"target_student_id"`
	TargetSubmissionID *uint           `json:"target_submission_id,omitempty"`
	Score              float64         `json:"score"`
	Feedback           string          `json:"feedback,omitempty"`
	RubricChecks       map[string]bool `json:"rubric_checks,omitempty"`
	AssignedBy         uint            `json:"assigned_by"`
	SubmittedAt        *time.Time      `json:"submitted_at,omitempty"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

type AssignmentGradeUpsertRequest struct {
	StudentID uint    `json:"student_id"`
	Score     float64 `json:"score"`
	Feedback  string  `json:"feedback"`
	Status    string  `json:"status"`
}

type AssignmentGradeBulkUpsertRequest struct {
	Items []AssignmentGradeUpsertRequest `json:"items"`
}

type AssignmentGradeResponse struct {
	ID           uint       `json:"id"`
	AssignmentID uint       `json:"assignment_id"`
	StudentID    uint       `json:"student_id"`
	MarkerID     uint       `json:"marker_id"`
	Score        float64    `json:"score"`
	Feedback     string     `json:"feedback,omitempty"`
	Status       string     `json:"status"`
	MarkedAt     *time.Time `json:"marked_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type QuizCreateRequest struct {
	CourseID        uint       `json:"course_id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	StartsAt        *time.Time `json:"starts_at"`
	EndsAt          *time.Time `json:"ends_at"`
	DurationMinutes uint       `json:"duration_minutes"`
	TotalMarks      float64    `json:"total_marks"`
	Status          string     `json:"status"`
}

type QuizUpdateRequest struct {
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	StartsAt        *time.Time `json:"starts_at"`
	EndsAt          *time.Time `json:"ends_at"`
	DurationMinutes uint       `json:"duration_minutes"`
	TotalMarks      float64    `json:"total_marks"`
	Status          string     `json:"status"`
}

type QuizResponse struct {
	ID              uint       `json:"id"`
	UUID            string     `json:"uuid"`
	CourseID        uint       `json:"course_id"`
	CourseCode      string     `json:"course_code,omitempty"`
	CourseTitle     string     `json:"course_title,omitempty"`
	Title           string     `json:"title"`
	Description     string     `json:"description,omitempty"`
	StartsAt        *time.Time `json:"starts_at,omitempty"`
	EndsAt          *time.Time `json:"ends_at,omitempty"`
	DurationMinutes uint       `json:"duration_minutes"`
	TotalMarks      float64    `json:"total_marks"`
	Status          string     `json:"status"`
	CreatedBy       uint       `json:"created_by"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type ExamCreateRequest struct {
	CourseID        uint           `json:"course_id"`
	CourseCode      string         `json:"course_code"`
	Title           string         `json:"title"`
	Description     string         `json:"description"`
	Instructions    string         `json:"instructions"`
	Venue           string         `json:"venue"`
	StartTime       time.Time      `json:"start_time"`
	EndTime         time.Time      `json:"end_time"`
	DurationMinutes uint           `json:"duration_minutes"`
	DeliveryMode    string         `json:"delivery_mode"`
	QuestionPayload map[string]any `json:"question_payload"`
	LecturerID      uint           `json:"lecturer_id"`
	ExamOfficerID   uint           `json:"exam_officer_id"`
	InvigilatorIDs  []uint         `json:"invigilator_ids"`
	Status          string         `json:"status"`
}

type ExamVenueCreateRequest struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	City     string `json:"city"`
	Country  string `json:"country"`
	Capacity uint   `json:"capacity"`
	IsActive *bool  `json:"is_active"`
}

type ExamVenueResponse struct {
	ID        uint      `json:"id"`
	UUID      string    `json:"uuid"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	City      string    `json:"city,omitempty"`
	Country   string    `json:"country"`
	Capacity  uint      `json:"capacity"`
	IsActive  bool      `json:"is_active"`
	CreatedBy uint      `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ExamStudentResponse struct {
	StudentID               uint   `json:"student_id"`
	Name                    string `json:"name"`
	MatricNo                string `json:"matric_no,omitempty"`
	Email                   string `json:"email,omitempty"`
	Level                   string `json:"level,omitempty"`
	AcademicSession         string `json:"academic_session,omitempty"`
	CountryOfResidence      string `json:"country_of_residence,omitempty"`
	IsInternational         bool   `json:"is_international"`
	RecommendedDeliveryMode string `json:"recommended_delivery_mode"`
}

type ExamVenueAllocationInput struct {
	StudentID    uint   `json:"student_id"`
	DeliveryMode string `json:"delivery_mode"`
	VenueID      *uint  `json:"venue_id"`
}

type ExamVenueAllocationRequest struct {
	Items []ExamVenueAllocationInput `json:"items"`
}

type ExamStudentAllocationResponse struct {
	ID                 uint      `json:"id"`
	UUID               string    `json:"uuid"`
	ExamID             uint      `json:"exam_id"`
	StudentID          uint      `json:"student_id"`
	DeliveryMode       string    `json:"delivery_mode"`
	VenueID            *uint     `json:"venue_id,omitempty"`
	VenueName          string    `json:"venue_name,omitempty"`
	VenueAddress       string    `json:"venue_address,omitempty"`
	IsInternational    bool      `json:"is_international"`
	CountryOfResidence string    `json:"country_of_residence,omitempty"`
	AllocatedBy        uint      `json:"allocated_by"`
	AllocatedAt        time.Time `json:"allocated_at"`
}

type ExamAttemptModerateRequest struct {
	Score   float64 `json:"score"`
	Comment string  `json:"comment"`
}

type ExamScriptPDFResponse struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	SizeBytes   int    `json:"size_bytes"`
}

type ExamUpdateRequest struct {
	Title           string         `json:"title"`
	Description     string         `json:"description"`
	Instructions    string         `json:"instructions"`
	Venue           string         `json:"venue"`
	StartTime       time.Time      `json:"start_time"`
	EndTime         time.Time      `json:"end_time"`
	DurationMinutes uint           `json:"duration_minutes"`
	DeliveryMode    string         `json:"delivery_mode"`
	QuestionPayload map[string]any `json:"question_payload"`
	ExamOfficerID   uint           `json:"exam_officer_id"`
	InvigilatorIDs  []uint         `json:"invigilator_ids"`
	Status          string         `json:"status"`
}

type ExamResponse struct {
	ID              uint                      `json:"id"`
	UUID            string                    `json:"uuid"`
	CourseID        uint                      `json:"course_id"`
	CourseCode      string                    `json:"course_code,omitempty"`
	CourseTitle     string                    `json:"course_title,omitempty"`
	Title           string                    `json:"title"`
	Description     string                    `json:"description,omitempty"`
	Instructions    string                    `json:"instructions,omitempty"`
	Venue           string                    `json:"venue,omitempty"`
	StartTime       time.Time                 `json:"start_time"`
	EndTime         time.Time                 `json:"end_time"`
	DurationMinutes uint                      `json:"duration_minutes"`
	DeliveryMode    string                    `json:"delivery_mode"`
	QuestionPayload map[string]any            `json:"question_payload,omitempty"`
	LecturerID      uint                      `json:"lecturer_id"`
	ExamOfficerID   uint                      `json:"exam_officer_id,omitempty"`
	ReleasedBy      uint                      `json:"released_by,omitempty"`
	ReleasedAt      *time.Time                `json:"released_at,omitempty"`
	Invigilators    []ExamInvigilatorResponse `json:"invigilators,omitempty"`
	Status          string                    `json:"status"`
	CreatedBy       uint                      `json:"created_by"`
	CreatedAt       time.Time                 `json:"created_at"`
	UpdatedAt       time.Time                 `json:"updated_at"`
}

type ExamInvigilatorResponse struct {
	ID            uint      `json:"id"`
	ExamID        uint      `json:"exam_id"`
	InvigilatorID uint      `json:"invigilator_id"`
	AssignedBy    uint      `json:"assigned_by"`
	AssignedAt    time.Time `json:"assigned_at"`
}

type ExamAttemptStartRequest struct {
	EnvironmentConfirmed bool `json:"environment_confirmed"`
}

type ExamAttemptSubmitRequest struct {
	AnswerPayload     map[string]any `json:"answer_payload"`
	IntegrityScore    int            `json:"integrity_score"`
	TerminationReason string         `json:"termination_reason"`
}

type ExamAttemptMarkRequest struct {
	Score    float64 `json:"score"`
	Feedback string  `json:"feedback"`
}

type ExamAttemptResponse struct {
	ID                     uint           `json:"id"`
	UUID                   string         `json:"uuid"`
	ExamID                 uint           `json:"exam_id"`
	StudentID              uint           `json:"student_id"`
	Status                 string         `json:"status"`
	QuestionPayload        map[string]any `json:"question_payload,omitempty"`
	AnswerPayload          map[string]any `json:"answer_payload,omitempty"`
	IntegrityScore         int            `json:"integrity_score"`
	EnvironmentConfirmedAt *time.Time     `json:"environment_confirmed_at,omitempty"`
	MonitoringArmedAt      *time.Time     `json:"monitoring_armed_at,omitempty"`
	SubmittedAt            *time.Time     `json:"submitted_at,omitempty"`
	SubmittedToOfficerAt   *time.Time     `json:"submitted_to_officer_at,omitempty"`
	SharedWithLecturerAt   *time.Time     `json:"shared_with_lecturer_at,omitempty"`
	MarkedBy               uint           `json:"marked_by,omitempty"`
	Score                  float64        `json:"score"`
	LecturerScore          float64        `json:"lecturer_score"`
	ModeratedScore         float64        `json:"moderated_score"`
	ModerationComment      string         `json:"moderation_comment,omitempty"`
	ModeratedBy            uint           `json:"moderated_by,omitempty"`
	ModeratedAt            *time.Time     `json:"moderated_at,omitempty"`
	Feedback               string         `json:"feedback,omitempty"`
	MarkedAt               *time.Time     `json:"marked_at,omitempty"`
	TerminationReason      string         `json:"termination_reason,omitempty"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
}

type ExamScriptAnnotationCreateRequest struct {
	QuestionID     string         `json:"question_id"`
	Comment        string         `json:"comment"`
	HighlightColor string         `json:"highlight_color"`
	InkPayload     map[string]any `json:"ink_payload"`
	Score          float64        `json:"score"`
}

type ExamScriptAnnotationResponse struct {
	ID             uint           `json:"id"`
	UUID           string         `json:"uuid"`
	AttemptID      uint           `json:"attempt_id"`
	LecturerID     uint           `json:"lecturer_id"`
	QuestionID     string         `json:"question_id"`
	Comment        string         `json:"comment,omitempty"`
	HighlightColor string         `json:"highlight_color,omitempty"`
	InkPayload     map[string]any `json:"ink_payload,omitempty"`
	Score          float64        `json:"score"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type ProctoringAlertCreateRequest struct {
	EventType      string         `json:"event_type"`
	Message        string         `json:"message"`
	Severity       string         `json:"severity"`
	IntegrityScore int            `json:"integrity_score"`
	Evidence       map[string]any `json:"evidence"`
}

type ProctoringAlertResponse struct {
	ID             uint           `json:"id"`
	UUID           string         `json:"uuid"`
	ExamID         uint           `json:"exam_id"`
	AttemptID      uint           `json:"attempt_id"`
	StudentID      uint           `json:"student_id"`
	InvigilatorID  *uint          `json:"invigilator_id,omitempty"`
	EventType      string         `json:"event_type"`
	Message        string         `json:"message"`
	Severity       string         `json:"severity"`
	IntegrityScore int            `json:"integrity_score"`
	Evidence       map[string]any `json:"evidence,omitempty"`
	AcknowledgedBy *uint          `json:"acknowledged_by,omitempty"`
	AcknowledgedAt *time.Time     `json:"acknowledged_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
}

type ResultCreateRequest struct {
	CourseID              uint       `json:"course_id"`
	CourseCode            string     `json:"course_code"`
	StudentID             uint       `json:"student_id"`
	AssessmentType        string     `json:"assessment_type"`
	ReferenceID           uint       `json:"reference_id"`
	Score                 float64    `json:"score"`
	GradedAssessmentScore float64    `json:"graded_assessment_score"`
	AssignmentScore       float64    `json:"assignment_score"`
	GroupAssignmentScore  float64    `json:"group_assignment_score"`
	PeerReviewScore       float64    `json:"peer_review_score"`
	ExaminationScore      float64    `json:"examination_score"`
	TotalScore            float64    `json:"total_score"`
	Grade                 string     `json:"grade"`
	Remark                string     `json:"remark"`
	Status                string     `json:"status"`
	PublishedAt           *time.Time `json:"published_at"`
}

type ResultUpdateRequest struct {
	Score                 float64    `json:"score"`
	GradedAssessmentScore float64    `json:"graded_assessment_score"`
	AssignmentScore       float64    `json:"assignment_score"`
	GroupAssignmentScore  float64    `json:"group_assignment_score"`
	PeerReviewScore       float64    `json:"peer_review_score"`
	ExaminationScore      float64    `json:"examination_score"`
	TotalScore            float64    `json:"total_score"`
	Grade                 string     `json:"grade"`
	Remark                string     `json:"remark"`
	Status                string     `json:"status"`
	PublishedAt           *time.Time `json:"published_at"`
}

type ResultResponse struct {
	ID                    uint       `json:"id"`
	UUID                  string     `json:"uuid"`
	CourseID              uint       `json:"course_id"`
	CourseCode            string     `json:"course_code,omitempty"`
	CourseTitle           string     `json:"course_title,omitempty"`
	StudentID             uint       `json:"student_id"`
	AssessmentType        string     `json:"assessment_type"`
	ReferenceID           uint       `json:"reference_id"`
	Score                 float64    `json:"score"`
	GradedAssessmentScore float64    `json:"graded_assessment_score"`
	AssignmentScore       float64    `json:"assignment_score"`
	GroupAssignmentScore  float64    `json:"group_assignment_score"`
	PeerReviewScore       float64    `json:"peer_review_score"`
	ExaminationScore      float64    `json:"examination_score"`
	TotalScore            float64    `json:"total_score"`
	Grade                 string     `json:"grade,omitempty"`
	Remark                string     `json:"remark,omitempty"`
	Status                string     `json:"status"`
	MarkedBy              uint       `json:"marked_by"`
	ApprovedBy            *uint      `json:"approved_by,omitempty"`
	ApprovedAt            *time.Time `json:"approved_at,omitempty"`
	PublishedAt           *time.Time `json:"published_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type CourseReportResponse struct {
	CourseID                  uint       `json:"course_id"`
	CourseCode                string     `json:"course_code,omitempty"`
	CourseTitle               string     `json:"course_title,omitempty"`
	MaterialCount             int64      `json:"material_count"`
	VideoLectureCount         int64      `json:"video_lecture_count"`
	LiveSessionCount          int64      `json:"live_session_count"`
	LiveSessionRecordingCount int64      `json:"live_session_recording_count"`
	AttendanceCount           int64      `json:"attendance_count"`
	PresentAttendanceCount    int64      `json:"present_attendance_count"`
	AssignmentCount           int64      `json:"assignment_count"`
	MarkedAssignmentCount     int64      `json:"marked_assignment_count"`
	QuizCount                 int64      `json:"quiz_count"`
	ExamCount                 int64      `json:"exam_count"`
	ResultCount               int64      `json:"result_count"`
	AverageResultScore        *float64   `json:"average_result_score,omitempty"`
	LastActivityAt            *time.Time `json:"last_activity_at,omitempty"`
}

type LecturerActivityItem struct {
	Type       string     `json:"type"`
	Summary    string     `json:"summary"`
	OccurredAt *time.Time `json:"occurred_at,omitempty"`
}

type LecturerCourseReportResponse struct {
	CourseID                  uint                   `json:"course_id"`
	CourseCode                string                 `json:"course_code,omitempty"`
	CourseTitle               string                 `json:"course_title,omitempty"`
	LecturerID                uint                   `json:"lecturer_id"`
	LecturerName              string                 `json:"lecturer_name,omitempty"`
	AssignmentsPublished      int64                  `json:"assignments_published"`
	AssignmentSubmissionCount int64                  `json:"assignment_submission_count"`
	AssignmentsMarked         int64                  `json:"assignments_marked"`
	AssignmentMarkingPercent  float64                `json:"assignment_marking_percent"`
	ForumDiscussionCount      int64                  `json:"forum_discussion_count"`
	ExamCount                 int64                  `json:"exam_count"`
	ExamMarkedCount           int64                  `json:"exam_marked_count"`
	AverageExamScore          *float64               `json:"average_exam_score,omitempty"`
	LiveSessionsScheduled     int64                  `json:"live_sessions_scheduled"`
	LiveSessionsConducted     int64                  `json:"live_sessions_conducted"`
	VideoLecturesUploaded     int64                  `json:"video_lectures_uploaded"`
	MaterialsUploaded         int64                  `json:"materials_uploaded"`
	HealthStatus              string                 `json:"health_status"`
	Activities                []LecturerActivityItem `json:"activities"`
}

type ExamWorkflowActionRequest struct {
	Comment string `json:"comment"`
}

type ExamScheduleRequest struct {
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	DurationMinutes uint      `json:"duration_minutes"`
	Venue           string    `json:"venue"`
	InvigilatorIDs  []uint    `json:"invigilator_ids"`
	Comment         string    `json:"comment"`
}

type LecturerExamScriptResponse struct {
	AttemptID            uint           `json:"attempt_id"`
	AttemptUUID          string         `json:"attempt_uuid"`
	ExamID               uint           `json:"exam_id"`
	ExamTitle            string         `json:"exam_title"`
	CourseID             uint           `json:"course_id"`
	CourseCode           string         `json:"course_code"`
	CourseTitle          string         `json:"course_title"`
	StudentID            uint           `json:"student_id"`
	StudentName          string         `json:"student_name"`
	CandidateNo          string         `json:"candidate_no"`
	Email                string         `json:"email,omitempty"`
	Status               string         `json:"status"`
	QuestionPayload      map[string]any `json:"question_payload,omitempty"`
	AnswerPayload        map[string]any `json:"answer_payload,omitempty"`
	QuestionTypes        []string       `json:"question_types"`
	QuestionCount        int            `json:"question_count"`
	ObjectiveScore       float64        `json:"objective_score"`
	TheoryScore          float64        `json:"theory_score"`
	PracticalScore       float64        `json:"practical_score"`
	Score                float64        `json:"score"`
	LecturerScore        float64        `json:"lecturer_score"`
	ModeratedScore       float64        `json:"moderated_score"`
	MaxScore             float64        `json:"max_score"`
	IntegrityScore       int            `json:"integrity_score"`
	Feedback             string         `json:"feedback,omitempty"`
	TerminationReason    string         `json:"termination_reason,omitempty"`
	SubmittedAt          *time.Time     `json:"submitted_at,omitempty"`
	SharedWithLecturerAt *time.Time     `json:"shared_with_lecturer_at,omitempty"`
}
