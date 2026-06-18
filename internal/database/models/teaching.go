package models

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CourseMaterialType string

const (
	CourseMaterialTypeDocument CourseMaterialType = "document"
	CourseMaterialTypeLink     CourseMaterialType = "link"
	CourseMaterialTypeArchive  CourseMaterialType = "archive"
	CourseMaterialTypeOther    CourseMaterialType = "other"
)

func (t CourseMaterialType) Valid() bool {
	switch t {
	case CourseMaterialTypeDocument, CourseMaterialTypeLink, CourseMaterialTypeArchive, CourseMaterialTypeOther:
		return true
	default:
		return false
	}
}

type VideoSourceType string

const (
	VideoSourceUploaded VideoSourceType = "uploaded"
	VideoSourceExternal VideoSourceType = "external"
)

func (t VideoSourceType) Valid() bool {
	switch t {
	case VideoSourceUploaded, VideoSourceExternal:
		return true
	default:
		return false
	}
}

type LiveSessionStatus string

const (
	LiveSessionStatusScheduled LiveSessionStatus = "scheduled"
	LiveSessionStatusLive      LiveSessionStatus = "live"
	LiveSessionStatusCompleted LiveSessionStatus = "completed"
	LiveSessionStatusCancelled LiveSessionStatus = "cancelled"
)

func (s LiveSessionStatus) Valid() bool {
	switch s {
	case LiveSessionStatusScheduled, LiveSessionStatusLive, LiveSessionStatusCompleted, LiveSessionStatusCancelled:
		return true
	default:
		return false
	}
}

type AttendanceStatus string

const (
	AttendanceStatusPresent AttendanceStatus = "present"
	AttendanceStatusAbsent  AttendanceStatus = "absent"
	AttendanceStatusLate    AttendanceStatus = "late"
	AttendanceStatusExcused AttendanceStatus = "excused"
)

func (s AttendanceStatus) Valid() bool {
	switch s {
	case AttendanceStatusPresent, AttendanceStatusAbsent, AttendanceStatusLate, AttendanceStatusExcused:
		return true
	default:
		return false
	}
}

type AssignmentStatus string

const (
	AssignmentStatusDraft     AssignmentStatus = "draft"
	AssignmentStatusPublished AssignmentStatus = "published"
	AssignmentStatusClosed    AssignmentStatus = "closed"
)

func (s AssignmentStatus) Valid() bool {
	switch s {
	case AssignmentStatusDraft, AssignmentStatusPublished, AssignmentStatusClosed:
		return true
	default:
		return false
	}
}

type AssignmentType string

const (
	AssignmentTypeIndividual AssignmentType = "individual"
	AssignmentTypeGroup      AssignmentType = "group"
	AssignmentTypePeerReview AssignmentType = "peer_review"
)

func (t AssignmentType) Valid() bool {
	switch t {
	case AssignmentTypeIndividual, AssignmentTypeGroup, AssignmentTypePeerReview:
		return true
	default:
		return false
	}
}

type AssignmentSubmissionMode string

const (
	AssignmentSubmissionModeText       AssignmentSubmissionMode = "text"
	AssignmentSubmissionModeFile       AssignmentSubmissionMode = "file"
	AssignmentSubmissionModeWhiteboard AssignmentSubmissionMode = "whiteboard"
	AssignmentSubmissionModeMixed      AssignmentSubmissionMode = "mixed"
)

func (m AssignmentSubmissionMode) Valid() bool {
	switch m {
	case AssignmentSubmissionModeText, AssignmentSubmissionModeFile, AssignmentSubmissionModeWhiteboard, AssignmentSubmissionModeMixed:
		return true
	default:
		return false
	}
}

type AssignmentGroupSource string

const (
	AssignmentGroupSourceRandomBackend      AssignmentGroupSource = "random_backend"
	AssignmentGroupSourceLecturerSetBackend AssignmentGroupSource = "lecturer_set_backend"
)

func (s AssignmentGroupSource) Valid() bool {
	switch s {
	case AssignmentGroupSourceRandomBackend, AssignmentGroupSourceLecturerSetBackend:
		return true
	default:
		return false
	}
}

type AssignmentSubmissionStatus string

const (
	AssignmentSubmissionStatusSubmitted   AssignmentSubmissionStatus = "submitted"
	AssignmentSubmissionStatusResubmitted AssignmentSubmissionStatus = "resubmitted"
	AssignmentSubmissionStatusMarked      AssignmentSubmissionStatus = "marked"
	AssignmentSubmissionStatusReturned    AssignmentSubmissionStatus = "returned"
)

func (s AssignmentSubmissionStatus) Valid() bool {
	switch s {
	case AssignmentSubmissionStatusSubmitted, AssignmentSubmissionStatusResubmitted, AssignmentSubmissionStatusMarked, AssignmentSubmissionStatusReturned:
		return true
	default:
		return false
	}
}

type QuizStatus string

const (
	QuizStatusDraft     QuizStatus = "draft"
	QuizStatusPublished QuizStatus = "published"
	QuizStatusClosed    QuizStatus = "closed"
)

func (s QuizStatus) Valid() bool {
	switch s {
	case QuizStatusDraft, QuizStatusPublished, QuizStatusClosed:
		return true
	default:
		return false
	}
}

type ExamStatus string

const (
	ExamStatusDraft         ExamStatus = "draft"
	ExamStatusOfficerReview ExamStatus = "officer_review"
	ExamStatusScheduled     ExamStatus = "scheduled"
	ExamStatusReleased      ExamStatus = "released"
	ExamStatusCompleted     ExamStatus = "completed"
	ExamStatusCancelled     ExamStatus = "cancelled"
)

func (s ExamStatus) Valid() bool {
	switch s {
	case ExamStatusDraft, ExamStatusOfficerReview, ExamStatusScheduled, ExamStatusReleased, ExamStatusCompleted, ExamStatusCancelled:
		return true
	default:
		return false
	}
}

type ExamDeliveryMode string

const (
	ExamDeliveryRemoteProctored ExamDeliveryMode = "remote_proctored"
	ExamDeliveryCenterBased     ExamDeliveryMode = "center_based"
)

func (m ExamDeliveryMode) Valid() bool {
	switch m {
	case ExamDeliveryRemoteProctored, ExamDeliveryCenterBased:
		return true
	default:
		return false
	}
}

type ExamAttemptStatus string

const (
	ExamAttemptInProgress          ExamAttemptStatus = "in_progress"
	ExamAttemptSubmitted           ExamAttemptStatus = "submitted"
	ExamAttemptSubmittedForMarking ExamAttemptStatus = "submitted_for_marking"
	ExamAttemptMarked              ExamAttemptStatus = "marked"
	ExamAttemptTerminated          ExamAttemptStatus = "terminated"
)

func (s ExamAttemptStatus) Valid() bool {
	switch s {
	case ExamAttemptInProgress, ExamAttemptSubmitted, ExamAttemptSubmittedForMarking, ExamAttemptMarked, ExamAttemptTerminated:
		return true
	default:
		return false
	}
}

type ProctoringAlertSeverity string

const (
	ProctoringAlertInfo     ProctoringAlertSeverity = "info"
	ProctoringAlertWarning  ProctoringAlertSeverity = "warning"
	ProctoringAlertCritical ProctoringAlertSeverity = "critical"
)

func (s ProctoringAlertSeverity) Valid() bool {
	switch s {
	case ProctoringAlertInfo, ProctoringAlertWarning, ProctoringAlertCritical:
		return true
	default:
		return false
	}
}

type ResultReferenceType string

const (
	ResultReferenceAssignment ResultReferenceType = "assignment"
	ResultReferenceQuiz       ResultReferenceType = "quiz"
	ResultReferenceExam       ResultReferenceType = "exam"
	ResultReferenceManual     ResultReferenceType = "manual"
)

func (t ResultReferenceType) Valid() bool {
	switch t {
	case ResultReferenceAssignment, ResultReferenceQuiz, ResultReferenceExam, ResultReferenceManual:
		return true
	default:
		return false
	}
}

type ResultStatus string

const (
	ResultStatusDraft     ResultStatus = "draft"
	ResultStatusSubmitted ResultStatus = "submitted"
	ResultStatusApproved  ResultStatus = "approved"
	ResultStatusPublished ResultStatus = "published"
)

func (s ResultStatus) Valid() bool {
	switch s {
	case ResultStatusDraft, ResultStatusSubmitted, ResultStatusApproved, ResultStatusPublished:
		return true
	default:
		return false
	}
}

type CourseMaterial struct {
	ID               uint               `gorm:"primaryKey"`
	UUID             string             `gorm:"size:36;uniqueIndex;not null"`
	CourseID         uint               `gorm:"not null;index"`
	Title            string             `gorm:"size:160;not null"`
	Description      string             `gorm:"type:text"`
	MaterialType     CourseMaterialType `gorm:"size:30;not null"`
	StoragePath      string             `gorm:"size:255"`
	OriginalFileName string             `gorm:"size:255"`
	StoredFileName   string             `gorm:"size:255;uniqueIndex"`
	ExternalURL      string             `gorm:"size:500"`
	MimeType         string             `gorm:"size:120"`
	SizeBytes        int64              `gorm:"not null;default:0"`
	AllowDownload    bool               `gorm:"not null;default:true"`
	UploadedBy       uint               `gorm:"not null;index"`
	PublishedAt      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time

	Course Course `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE"`
}

func (m *CourseMaterial) BeforeCreate(_ *gorm.DB) error {
	if strings.TrimSpace(m.UUID) == "" {
		m.UUID = uuid.NewString()
	}

	return nil
}

func (m *CourseMaterial) BeforeSave(_ *gorm.DB) error {
	m.Title = strings.TrimSpace(m.Title)
	m.Description = strings.TrimSpace(m.Description)
	m.StoragePath = strings.TrimSpace(m.StoragePath)
	m.OriginalFileName = strings.TrimSpace(m.OriginalFileName)
	m.StoredFileName = strings.TrimSpace(m.StoredFileName)
	m.ExternalURL = strings.TrimSpace(m.ExternalURL)
	m.MimeType = strings.TrimSpace(m.MimeType)

	if m.CourseID == 0 || m.UploadedBy == 0 || m.Title == "" {
		return errors.New("course_id, uploaded_by, and title are required")
	}

	if !m.MaterialType.Valid() {
		return errors.New("invalid material type")
	}

	if m.MaterialType == CourseMaterialTypeLink {
		if m.ExternalURL == "" {
			return errors.New("external_url is required for link materials")
		}
	} else if m.StoragePath == "" && m.ExternalURL == "" {
		return errors.New("storage_path or external_url is required")
	}

	return nil
}

type VideoLecture struct {
	ID                 uint            `gorm:"primaryKey"`
	UUID               string          `gorm:"size:36;uniqueIndex;not null"`
	CourseID           uint            `gorm:"not null;index"`
	Title              string          `gorm:"size:180;not null"`
	Subtitle           string          `gorm:"size:180"`
	Description        string          `gorm:"type:text"`
	LecturerName       string          `gorm:"size:140"`
	SourceType         VideoSourceType `gorm:"size:20;not null"`
	StoragePath        string          `gorm:"size:255"`
	OriginalFileName   string          `gorm:"size:255"`
	StoredFileName     string          `gorm:"size:255;uniqueIndex"`
	ExternalURL        string          `gorm:"size:500"`
	MimeType           string          `gorm:"size:120"`
	SizeBytes          int64           `gorm:"not null;default:0"`
	DurationMinutes    uint            `gorm:"not null;default:0"`
	AudienceKeys       []string        `gorm:"serializer:json"`
	Tags               []string        `gorm:"serializer:json"`
	AllowDownload      bool            `gorm:"not null;default:true"`
	RequireWatchedMark bool            `gorm:"not null;default:false"`
	UploadedBy         uint            `gorm:"not null;index"`
	PublishedAt        time.Time       `gorm:"not null"`
	CreatedAt          time.Time
	UpdatedAt          time.Time

	Course  Course              `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE"`
	Watches []VideoLectureWatch `gorm:"constraint:OnDelete:CASCADE"`
}

func (v *VideoLecture) BeforeCreate(_ *gorm.DB) error {
	if strings.TrimSpace(v.UUID) == "" {
		v.UUID = uuid.NewString()
	}

	if v.PublishedAt.IsZero() {
		v.PublishedAt = time.Now().UTC()
	}

	return nil
}

func (v *VideoLecture) BeforeSave(_ *gorm.DB) error {
	v.Title = strings.TrimSpace(v.Title)
	v.Subtitle = strings.TrimSpace(v.Subtitle)
	v.Description = strings.TrimSpace(v.Description)
	v.LecturerName = strings.TrimSpace(v.LecturerName)
	v.StoragePath = strings.TrimSpace(v.StoragePath)
	v.OriginalFileName = strings.TrimSpace(v.OriginalFileName)
	v.StoredFileName = strings.TrimSpace(v.StoredFileName)
	v.ExternalURL = strings.TrimSpace(v.ExternalURL)
	v.MimeType = strings.TrimSpace(v.MimeType)

	if v.CourseID == 0 || v.UploadedBy == 0 || v.Title == "" {
		return errors.New("course_id, uploaded_by, and title are required")
	}

	if !v.SourceType.Valid() {
		return errors.New("invalid video source type")
	}

	if v.SourceType == VideoSourceUploaded && v.StoragePath == "" {
		return errors.New("storage_path is required for uploaded video lectures")
	}

	if v.SourceType == VideoSourceExternal && v.ExternalURL == "" {
		return errors.New("external_url is required for external video lectures")
	}

	return nil
}

type VideoLectureWatch struct {
	ID             uint      `gorm:"primaryKey"`
	VideoLectureID uint      `gorm:"not null;uniqueIndex:idx_video_lecture_watch"`
	UserID         uint      `gorm:"not null;uniqueIndex:idx_video_lecture_watch"`
	WatchedAt      time.Time `gorm:"not null"`
	CreatedAt      time.Time

	VideoLecture VideoLecture `gorm:"foreignKey:VideoLectureID;constraint:OnDelete:CASCADE"`
}

func (w *VideoLectureWatch) BeforeSave(_ *gorm.DB) error {
	if w.VideoLectureID == 0 || w.UserID == 0 {
		return errors.New("video_lecture_id and user_id are required")
	}

	if w.WatchedAt.IsZero() {
		w.WatchedAt = time.Now().UTC()
	}

	return nil
}

type LiveSessionSettings struct {
	StudentCameraRequired  bool `json:"student_camera_required"`
	CaptureRegistrationNo  bool `json:"capture_registration_number"`
	AllowStudentRecording  bool `json:"allow_student_recording"`
	AllowLecturerRecording bool `json:"allow_lecturer_recording"`
	AttendanceEnabled      bool `json:"attendance_enabled"`
	ChatEnabled            bool `json:"chat_enabled"`
	QuestionsEnabled       bool `json:"questions_enabled"`
}

type LiveSession struct {
	ID           uint                `gorm:"primaryKey"`
	UUID         string              `gorm:"size:36;uniqueIndex;not null"`
	CourseID     uint                `gorm:"not null;index"`
	Title        string              `gorm:"size:180;not null"`
	Description  string              `gorm:"type:text"`
	LecturerName string              `gorm:"size:140"`
	RoomName     string              `gorm:"size:120;not null"`
	StartTime    time.Time           `gorm:"not null;index"`
	EndTime      time.Time           `gorm:"not null;index"`
	Status       LiveSessionStatus   `gorm:"size:20;not null;default:scheduled"`
	Agenda       []string            `gorm:"serializer:json"`
	Materials    []string            `gorm:"serializer:json"`
	Settings     LiveSessionSettings `gorm:"serializer:json"`
	CreatedBy    uint                `gorm:"not null;index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time

	Course     Course                  `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE"`
	Attendance []LiveSessionAttendance `gorm:"constraint:OnDelete:CASCADE"`
	Recordings []LiveSessionRecording  `gorm:"constraint:OnDelete:CASCADE"`
}

func (s *LiveSession) BeforeCreate(_ *gorm.DB) error {
	if strings.TrimSpace(s.UUID) == "" {
		s.UUID = uuid.NewString()
	}

	return nil
}

func (s *LiveSession) BeforeSave(_ *gorm.DB) error {
	s.Title = strings.TrimSpace(s.Title)
	s.Description = strings.TrimSpace(s.Description)
	s.LecturerName = strings.TrimSpace(s.LecturerName)
	s.RoomName = strings.TrimSpace(s.RoomName)

	if s.CourseID == 0 || s.CreatedBy == 0 || s.Title == "" || s.RoomName == "" {
		return errors.New("course_id, created_by, title, and room_name are required")
	}

	if !s.Status.Valid() {
		return errors.New("invalid live session status")
	}

	if s.StartTime.IsZero() || s.EndTime.IsZero() || !s.EndTime.After(s.StartTime) {
		return errors.New("start_time and end_time must be valid and end_time must be after start_time")
	}

	return nil
}

type LiveSessionAttendance struct {
	ID                 uint             `gorm:"primaryKey"`
	LiveSessionID      uint             `gorm:"not null;uniqueIndex:idx_live_session_attendance"`
	UserID             uint             `gorm:"not null;uniqueIndex:idx_live_session_attendance"`
	RegistrationNumber string           `gorm:"size:50"`
	Status             AttendanceStatus `gorm:"size:20;not null;default:present"`
	JoinedAt           *time.Time
	LeftAt             *time.Time
	DurationMinutes    uint `gorm:"not null;default:0"`
	CapturedBy         uint `gorm:"not null;index"`
	CreatedAt          time.Time
	UpdatedAt          time.Time

	LiveSession LiveSession `gorm:"foreignKey:LiveSessionID;constraint:OnDelete:CASCADE"`
}

func (a *LiveSessionAttendance) BeforeSave(_ *gorm.DB) error {
	a.RegistrationNumber = strings.TrimSpace(a.RegistrationNumber)

	if a.LiveSessionID == 0 || a.UserID == 0 || a.CapturedBy == 0 {
		return errors.New("live_session_id, user_id, and captured_by are required")
	}

	if !a.Status.Valid() {
		return errors.New("invalid attendance status")
	}

	return nil
}

type LiveSessionRecording struct {
	ID               uint            `gorm:"primaryKey"`
	UUID             string          `gorm:"size:36;uniqueIndex;not null"`
	LiveSessionID    uint            `gorm:"not null;index"`
	Title            string          `gorm:"size:180;not null"`
	Description      string          `gorm:"type:text"`
	SourceType       VideoSourceType `gorm:"size:20;not null"`
	StoragePath      string          `gorm:"size:255"`
	OriginalFileName string          `gorm:"size:255"`
	StoredFileName   string          `gorm:"size:255;uniqueIndex"`
	ExternalURL      string          `gorm:"size:500"`
	MimeType         string          `gorm:"size:120"`
	SizeBytes        int64           `gorm:"not null;default:0"`
	AddedBy          uint            `gorm:"not null;index"`
	PublishedAt      time.Time       `gorm:"not null"`
	CreatedAt        time.Time
	UpdatedAt        time.Time

	LiveSession LiveSession `gorm:"foreignKey:LiveSessionID;constraint:OnDelete:CASCADE"`
}

func (r *LiveSessionRecording) BeforeCreate(_ *gorm.DB) error {
	if strings.TrimSpace(r.UUID) == "" {
		r.UUID = uuid.NewString()
	}

	if r.PublishedAt.IsZero() {
		r.PublishedAt = time.Now().UTC()
	}

	return nil
}

func (r *LiveSessionRecording) BeforeSave(_ *gorm.DB) error {
	r.Title = strings.TrimSpace(r.Title)
	r.Description = strings.TrimSpace(r.Description)
	r.StoragePath = strings.TrimSpace(r.StoragePath)
	r.OriginalFileName = strings.TrimSpace(r.OriginalFileName)
	r.StoredFileName = strings.TrimSpace(r.StoredFileName)
	r.ExternalURL = strings.TrimSpace(r.ExternalURL)
	r.MimeType = strings.TrimSpace(r.MimeType)

	if r.LiveSessionID == 0 || r.AddedBy == 0 || r.Title == "" {
		return errors.New("live_session_id, added_by, and title are required")
	}

	if !r.SourceType.Valid() {
		return errors.New("invalid recording source type")
	}

	if r.SourceType == VideoSourceUploaded && r.StoragePath == "" {
		return errors.New("storage_path is required for uploaded recordings")
	}

	if r.SourceType == VideoSourceExternal && r.ExternalURL == "" {
		return errors.New("external_url is required for external recordings")
	}

	return nil
}

type CourseForumPost struct {
	ID        uint   `gorm:"primaryKey"`
	UUID      string `gorm:"size:36;uniqueIndex;not null"`
	CourseID  uint   `gorm:"not null;index"`
	AuthorID  uint   `gorm:"not null;index"`
	Title     string `gorm:"size:180"`
	Body      string `gorm:"type:text;not null"`
	IsPinned  bool   `gorm:"not null;default:false"`
	IsLocked  bool   `gorm:"not null;default:false"`
	ParentID  *uint  `gorm:"index"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Course  Course            `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE"`
	Author  User              `gorm:"foreignKey:AuthorID;constraint:OnDelete:CASCADE"`
	Parent  *CourseForumPost  `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE"`
	Replies []CourseForumPost `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE"`
}

func (p *CourseForumPost) BeforeCreate(_ *gorm.DB) error {
	if strings.TrimSpace(p.UUID) == "" {
		p.UUID = uuid.NewString()
	}

	return nil
}

func (p *CourseForumPost) BeforeSave(_ *gorm.DB) error {
	p.Title = strings.TrimSpace(p.Title)
	p.Body = strings.TrimSpace(p.Body)

	if p.CourseID == 0 || p.AuthorID == 0 || p.Body == "" {
		return errors.New("course_id, author_id, and body are required")
	}

	return nil
}

type CourseDirectMessage struct {
	ID          uint       `gorm:"primaryKey"`
	UUID        string     `gorm:"size:36;uniqueIndex;not null"`
	CourseID    uint       `gorm:"not null;index"`
	SenderID    uint       `gorm:"not null;index"`
	RecipientID uint       `gorm:"not null;index"`
	Body        string     `gorm:"type:text;not null"`
	ReadAt      *time.Time `gorm:"index"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Course    Course `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE"`
	Sender    User   `gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE"`
	Recipient User   `gorm:"foreignKey:RecipientID;constraint:OnDelete:CASCADE"`
}

func (m *CourseDirectMessage) BeforeCreate(_ *gorm.DB) error {
	if strings.TrimSpace(m.UUID) == "" {
		m.UUID = uuid.NewString()
	}
	return nil
}

func (m *CourseDirectMessage) BeforeSave(_ *gorm.DB) error {
	m.Body = strings.TrimSpace(m.Body)
	if m.CourseID == 0 || m.SenderID == 0 || m.RecipientID == 0 || m.Body == "" {
		return errors.New("course_id, sender_id, recipient_id, and body are required")
	}
	if m.SenderID == m.RecipientID {
		return errors.New("sender and recipient must be different")
	}
	return nil
}

type Assignment struct {
	ID                 uint                     `gorm:"primaryKey"`
	UUID               string                   `gorm:"size:36;uniqueIndex;not null"`
	CourseID           uint                     `gorm:"not null;index"`
	Title              string                   `gorm:"size:180;not null"`
	Description        string                   `gorm:"type:text"`
	Instructions       string                   `gorm:"type:text"`
	DueAt              *time.Time               `gorm:"index"`
	MaxScore           float64                  `gorm:"not null"`
	AssignmentType     AssignmentType           `gorm:"size:30;not null;default:individual"`
	SubmissionMode     AssignmentSubmissionMode `gorm:"size:30;not null;default:mixed"`
	AllowedExtensions  []string                 `gorm:"serializer:json"`
	WhiteboardEnabled  bool                     `gorm:"not null;default:false"`
	WhiteboardRequired bool                     `gorm:"not null;default:false"`
	WhiteboardPrompt   string                   `gorm:"type:text"`
	PeerReviewEnabled  bool                     `gorm:"not null;default:false"`
	PeerReviewRubric   []string                 `gorm:"serializer:json"`
	GroupSource        AssignmentGroupSource    `gorm:"size:30"`
	Status             AssignmentStatus         `gorm:"size:20;not null;default:draft"`
	CreatedBy          uint                     `gorm:"not null;index"`
	CreatedAt          time.Time
	UpdatedAt          time.Time

	Course      Course                 `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE"`
	Groups      []AssignmentGroup      `gorm:"constraint:OnDelete:CASCADE"`
	Submissions []AssignmentSubmission `gorm:"constraint:OnDelete:CASCADE"`
	PeerReviews []AssignmentPeerReview `gorm:"constraint:OnDelete:CASCADE"`
	Grades      []AssignmentGrade      `gorm:"constraint:OnDelete:CASCADE"`
}

func (a *Assignment) BeforeCreate(_ *gorm.DB) error {
	if strings.TrimSpace(a.UUID) == "" {
		a.UUID = uuid.NewString()
	}

	return nil
}

func (a *Assignment) BeforeSave(_ *gorm.DB) error {
	a.Title = strings.TrimSpace(a.Title)
	a.Description = strings.TrimSpace(a.Description)
	a.Instructions = strings.TrimSpace(a.Instructions)
	a.WhiteboardPrompt = strings.TrimSpace(a.WhiteboardPrompt)

	if a.CourseID == 0 || a.CreatedBy == 0 || a.Title == "" {
		return errors.New("course_id, created_by, and title are required")
	}

	if a.AssignmentType == "" {
		a.AssignmentType = AssignmentTypeIndividual
	}
	if !a.AssignmentType.Valid() {
		return errors.New("invalid assignment type")
	}

	if a.SubmissionMode == "" {
		a.SubmissionMode = AssignmentSubmissionModeMixed
	}
	if !a.SubmissionMode.Valid() {
		return errors.New("invalid assignment submission mode")
	}

	if a.AssignmentType == AssignmentTypeGroup && a.GroupSource != "" && !a.GroupSource.Valid() {
		return errors.New("invalid assignment group source")
	}

	if !a.Status.Valid() {
		return errors.New("invalid assignment status")
	}

	if a.MaxScore <= 0 {
		return errors.New("max_score must be greater than zero")
	}

	return nil
}

type AssignmentGroup struct {
	ID           uint                  `gorm:"primaryKey"`
	UUID         string                `gorm:"size:36;uniqueIndex;not null"`
	AssignmentID uint                  `gorm:"not null;index"`
	Name         string                `gorm:"size:140;not null"`
	Source       AssignmentGroupSource `gorm:"size:30;not null"`
	CreatedBy    uint                  `gorm:"not null;index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time

	Assignment Assignment              `gorm:"foreignKey:AssignmentID;constraint:OnDelete:CASCADE"`
	Members    []AssignmentGroupMember `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE"`
}

func (g *AssignmentGroup) BeforeCreate(_ *gorm.DB) error {
	if strings.TrimSpace(g.UUID) == "" {
		g.UUID = uuid.NewString()
	}
	return nil
}

func (g *AssignmentGroup) BeforeSave(_ *gorm.DB) error {
	g.Name = strings.TrimSpace(g.Name)
	if g.AssignmentID == 0 || g.CreatedBy == 0 || g.Name == "" {
		return errors.New("assignment_id, created_by, and name are required")
	}
	if g.Source == "" {
		g.Source = AssignmentGroupSourceLecturerSetBackend
	}
	if !g.Source.Valid() {
		return errors.New("invalid assignment group source")
	}
	return nil
}

type AssignmentGroupMember struct {
	ID        uint `gorm:"primaryKey"`
	GroupID   uint `gorm:"not null;uniqueIndex:idx_assignment_group_member"`
	StudentID uint `gorm:"not null;uniqueIndex:idx_assignment_group_member"`
	AddedBy   uint `gorm:"not null;index"`
	CreatedAt time.Time

	Group   AssignmentGroup `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE"`
	Student User            `gorm:"foreignKey:StudentID;constraint:OnDelete:CASCADE"`
}

func (m *AssignmentGroupMember) BeforeSave(_ *gorm.DB) error {
	if m.GroupID == 0 || m.StudentID == 0 || m.AddedBy == 0 {
		return errors.New("group_id, student_id, and added_by are required")
	}
	return nil
}

type AssignmentSubmission struct {
	ID             uint                       `gorm:"primaryKey"`
	UUID           string                     `gorm:"size:36;uniqueIndex;not null"`
	AssignmentID   uint                       `gorm:"not null;uniqueIndex:idx_assignment_submission_owner"`
	StudentID      uint                       `gorm:"not null;uniqueIndex:idx_assignment_submission_owner"`
	GroupID        *uint                      `gorm:"uniqueIndex:idx_assignment_submission_owner"`
	TextAnswer     string                     `gorm:"type:text"`
	WhiteboardData []byte                     `gorm:"type:bytea"`
	Status         AssignmentSubmissionStatus `gorm:"size:30;not null;default:submitted"`
	SubmittedAt    time.Time                  `gorm:"not null;index"`
	CreatedAt      time.Time
	UpdatedAt      time.Time

	Assignment Assignment                 `gorm:"foreignKey:AssignmentID;constraint:OnDelete:CASCADE"`
	Student    User                       `gorm:"foreignKey:StudentID;constraint:OnDelete:CASCADE"`
	Group      *AssignmentGroup           `gorm:"foreignKey:GroupID;constraint:OnDelete:SET NULL"`
	Files      []AssignmentSubmissionFile `gorm:"foreignKey:SubmissionID;constraint:OnDelete:CASCADE"`
}

func (s *AssignmentSubmission) BeforeCreate(_ *gorm.DB) error {
	if strings.TrimSpace(s.UUID) == "" {
		s.UUID = uuid.NewString()
	}
	if s.SubmittedAt.IsZero() {
		s.SubmittedAt = time.Now().UTC()
	}
	return nil
}

func (s *AssignmentSubmission) BeforeSave(_ *gorm.DB) error {
	s.TextAnswer = strings.TrimSpace(s.TextAnswer)
	if s.AssignmentID == 0 || s.StudentID == 0 {
		return errors.New("assignment_id and student_id are required")
	}
	if s.Status == "" {
		s.Status = AssignmentSubmissionStatusSubmitted
	}
	if !s.Status.Valid() {
		return errors.New("invalid assignment submission status")
	}
	return nil
}

type AssignmentSubmissionFile struct {
	ID               uint   `gorm:"primaryKey"`
	SubmissionID     uint   `gorm:"not null;index"`
	StoragePath      string `gorm:"size:255;not null"`
	OriginalFileName string `gorm:"size:255;not null"`
	StoredFileName   string `gorm:"size:255;uniqueIndex;not null"`
	MimeType         string `gorm:"size:120"`
	SizeBytes        int64  `gorm:"not null;default:0"`
	CreatedAt        time.Time

	Submission AssignmentSubmission `gorm:"foreignKey:SubmissionID;constraint:OnDelete:CASCADE"`
}

func (f *AssignmentSubmissionFile) BeforeSave(_ *gorm.DB) error {
	f.StoragePath = strings.TrimSpace(f.StoragePath)
	f.OriginalFileName = strings.TrimSpace(f.OriginalFileName)
	f.StoredFileName = strings.TrimSpace(f.StoredFileName)
	f.MimeType = strings.TrimSpace(f.MimeType)
	if f.SubmissionID == 0 || f.StoragePath == "" || f.OriginalFileName == "" || f.StoredFileName == "" {
		return errors.New("submission_id, storage_path, original_file_name, and stored_file_name are required")
	}
	return nil
}

type AssignmentPeerReview struct {
	ID                 uint            `gorm:"primaryKey"`
	UUID               string          `gorm:"size:36;uniqueIndex;not null"`
	AssignmentID       uint            `gorm:"not null;uniqueIndex:idx_assignment_peer_review"`
	ReviewerID         uint            `gorm:"not null;uniqueIndex:idx_assignment_peer_review"`
	TargetStudentID    uint            `gorm:"not null;index"`
	TargetSubmissionID *uint           `gorm:"index"`
	Score              float64         `gorm:"not null;default:0"`
	Feedback           string          `gorm:"type:text"`
	RubricChecks       map[string]bool `gorm:"serializer:json"`
	AssignedBy         uint            `gorm:"not null;index"`
	SubmittedAt        *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time

	Assignment       Assignment            `gorm:"foreignKey:AssignmentID;constraint:OnDelete:CASCADE"`
	Reviewer         User                  `gorm:"foreignKey:ReviewerID;constraint:OnDelete:CASCADE"`
	TargetStudent    User                  `gorm:"foreignKey:TargetStudentID;constraint:OnDelete:CASCADE"`
	TargetSubmission *AssignmentSubmission `gorm:"foreignKey:TargetSubmissionID;constraint:OnDelete:SET NULL"`
}

func (p *AssignmentPeerReview) BeforeCreate(_ *gorm.DB) error {
	if strings.TrimSpace(p.UUID) == "" {
		p.UUID = uuid.NewString()
	}
	return nil
}

func (p *AssignmentPeerReview) BeforeSave(_ *gorm.DB) error {
	p.Feedback = strings.TrimSpace(p.Feedback)
	if p.AssignmentID == 0 || p.ReviewerID == 0 || p.TargetStudentID == 0 || p.AssignedBy == 0 {
		return errors.New("assignment_id, reviewer_id, target_student_id, and assigned_by are required")
	}
	return nil
}

type AssignmentGrade struct {
	ID           uint    `gorm:"primaryKey"`
	AssignmentID uint    `gorm:"not null;uniqueIndex:idx_assignment_grade_student"`
	StudentID    uint    `gorm:"not null;uniqueIndex:idx_assignment_grade_student"`
	MarkerID     uint    `gorm:"not null;index"`
	Score        float64 `gorm:"not null"`
	Feedback     string  `gorm:"type:text"`
	Status       string  `gorm:"size:20;not null;default:marked"`
	MarkedAt     *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time

	Assignment Assignment `gorm:"foreignKey:AssignmentID;constraint:OnDelete:CASCADE"`
}

func (g *AssignmentGrade) BeforeSave(_ *gorm.DB) error {
	g.Feedback = strings.TrimSpace(g.Feedback)
	g.Status = strings.ToLower(strings.TrimSpace(g.Status))

	if g.AssignmentID == 0 || g.StudentID == 0 || g.MarkerID == 0 {
		return errors.New("assignment_id, student_id, and marker_id are required")
	}

	if g.Status == "" {
		g.Status = "marked"
	}

	return nil
}

type Quiz struct {
	ID              uint       `gorm:"primaryKey"`
	UUID            string     `gorm:"size:36;uniqueIndex;not null"`
	CourseID        uint       `gorm:"not null;index"`
	Title           string     `gorm:"size:180;not null"`
	Description     string     `gorm:"type:text"`
	StartsAt        *time.Time `gorm:"index"`
	EndsAt          *time.Time `gorm:"index"`
	DurationMinutes uint       `gorm:"not null;default:0"`
	TotalMarks      float64    `gorm:"not null"`
	Status          QuizStatus `gorm:"size:20;not null;default:draft"`
	CreatedBy       uint       `gorm:"not null;index"`
	CreatedAt       time.Time
	UpdatedAt       time.Time

	Course Course `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE"`
}

func (q *Quiz) BeforeCreate(_ *gorm.DB) error {
	if strings.TrimSpace(q.UUID) == "" {
		q.UUID = uuid.NewString()
	}

	return nil
}

func (q *Quiz) BeforeSave(_ *gorm.DB) error {
	q.Title = strings.TrimSpace(q.Title)
	q.Description = strings.TrimSpace(q.Description)

	if q.CourseID == 0 || q.CreatedBy == 0 || q.Title == "" {
		return errors.New("course_id, created_by, and title are required")
	}

	if !q.Status.Valid() {
		return errors.New("invalid quiz status")
	}

	if q.TotalMarks <= 0 {
		return errors.New("total_marks must be greater than zero")
	}

	if q.StartsAt != nil && q.EndsAt != nil && !q.EndsAt.After(*q.StartsAt) {
		return errors.New("ends_at must be after starts_at")
	}

	return nil
}

type Exam struct {
	ID              uint             `gorm:"primaryKey"`
	UUID            string           `gorm:"size:36;uniqueIndex;not null"`
	CourseID        uint             `gorm:"not null;index"`
	Title           string           `gorm:"size:180;not null"`
	Description     string           `gorm:"type:text"`
	Instructions    string           `gorm:"type:text"`
	Venue           string           `gorm:"size:140"`
	StartTime       time.Time        `gorm:"not null;index"`
	EndTime         time.Time        `gorm:"not null;index"`
	DurationMinutes uint             `gorm:"not null;default:0"`
	Status          ExamStatus       `gorm:"size:30;not null;default:draft"`
	DeliveryMode    ExamDeliveryMode `gorm:"size:30;not null;default:remote_proctored"`
	QuestionPayload []byte           `gorm:"type:bytea"`
	LecturerID      uint             `gorm:"not null;index"`
	ExamOfficerID   uint             `gorm:"index"`
	ReleasedBy      uint             `gorm:"index"`
	ReleasedAt      *time.Time
	CreatedBy       uint `gorm:"not null;index"`
	CreatedAt       time.Time
	UpdatedAt       time.Time

	Course       Course                  `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE"`
	Invigilators []ExamInvigilator       `gorm:"constraint:OnDelete:CASCADE"`
	Attempts     []ExamAttempt           `gorm:"constraint:OnDelete:CASCADE"`
	Allocations  []ExamStudentAllocation `gorm:"constraint:OnDelete:CASCADE"`
}

func (e *Exam) BeforeCreate(_ *gorm.DB) error {
	if strings.TrimSpace(e.UUID) == "" {
		e.UUID = uuid.NewString()
	}

	return nil
}

func (e *Exam) BeforeSave(_ *gorm.DB) error {
	e.Title = strings.TrimSpace(e.Title)
	e.Description = strings.TrimSpace(e.Description)
	e.Instructions = strings.TrimSpace(e.Instructions)
	e.Venue = strings.TrimSpace(e.Venue)

	if e.CourseID == 0 || e.CreatedBy == 0 || e.Title == "" {
		return errors.New("course_id, created_by, and title are required")
	}
	if e.LecturerID == 0 {
		e.LecturerID = e.CreatedBy
	}

	if !e.Status.Valid() {
		return errors.New("invalid exam status")
	}
	if e.DeliveryMode == "" {
		e.DeliveryMode = ExamDeliveryRemoteProctored
	}
	if !e.DeliveryMode.Valid() {
		return errors.New("invalid exam delivery mode")
	}

	if e.StartTime.IsZero() || e.EndTime.IsZero() || !e.EndTime.After(e.StartTime) {
		return errors.New("start_time and end_time must be valid and end_time must be after start_time")
	}

	return nil
}

type ExamVenue struct {
	ID        uint   `gorm:"primaryKey"`
	UUID      string `gorm:"size:36;uniqueIndex;not null"`
	Name      string `gorm:"size:160;not null"`
	Address   string `gorm:"size:300;not null"`
	City      string `gorm:"size:100"`
	Country   string `gorm:"size:100;not null;default:Nigeria"`
	Capacity  uint   `gorm:"not null;default:0"`
	IsActive  bool   `gorm:"not null;default:true;index"`
	CreatedBy uint   `gorm:"not null;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (v *ExamVenue) BeforeCreate(_ *gorm.DB) error {
	if strings.TrimSpace(v.UUID) == "" {
		v.UUID = uuid.NewString()
	}
	return nil
}

func (v *ExamVenue) BeforeSave(_ *gorm.DB) error {
	v.Name = strings.TrimSpace(v.Name)
	v.Address = strings.TrimSpace(v.Address)
	v.City = strings.TrimSpace(v.City)
	v.Country = strings.TrimSpace(v.Country)
	if v.Country == "" {
		v.Country = "Nigeria"
	}
	if v.Name == "" || v.Address == "" || v.CreatedBy == 0 {
		return errors.New("name, address, and created_by are required")
	}
	return nil
}

type ExamStudentAllocation struct {
	ID                 uint             `gorm:"primaryKey"`
	UUID               string           `gorm:"size:36;uniqueIndex;not null"`
	ExamID             uint             `gorm:"not null;uniqueIndex:idx_exam_student_allocation"`
	StudentID          uint             `gorm:"not null;uniqueIndex:idx_exam_student_allocation"`
	DeliveryMode       ExamDeliveryMode `gorm:"size:30;not null"`
	VenueID            *uint            `gorm:"index"`
	VenueName          string           `gorm:"size:160"`
	VenueAddress       string           `gorm:"size:300"`
	IsInternational    bool             `gorm:"not null;default:false;index"`
	CountryOfResidence string           `gorm:"size:80"`
	AllocatedBy        uint             `gorm:"not null;index"`
	AllocatedAt        time.Time        `gorm:"not null"`
	CreatedAt          time.Time
	UpdatedAt          time.Time

	Exam    Exam       `gorm:"foreignKey:ExamID;constraint:OnDelete:CASCADE"`
	Student User       `gorm:"foreignKey:StudentID;constraint:OnDelete:CASCADE"`
	Venue   *ExamVenue `gorm:"foreignKey:VenueID;constraint:OnDelete:SET NULL"`
}

func (a *ExamStudentAllocation) BeforeCreate(_ *gorm.DB) error {
	if strings.TrimSpace(a.UUID) == "" {
		a.UUID = uuid.NewString()
	}
	return nil
}

func (a *ExamStudentAllocation) BeforeSave(_ *gorm.DB) error {
	a.VenueName = strings.TrimSpace(a.VenueName)
	a.VenueAddress = strings.TrimSpace(a.VenueAddress)
	a.CountryOfResidence = strings.TrimSpace(a.CountryOfResidence)
	if a.ExamID == 0 || a.StudentID == 0 || a.AllocatedBy == 0 {
		return errors.New("exam_id, student_id, and allocated_by are required")
	}
	if a.DeliveryMode == "" {
		a.DeliveryMode = ExamDeliveryRemoteProctored
	}
	if !a.DeliveryMode.Valid() {
		return errors.New("invalid exam delivery mode")
	}
	if a.DeliveryMode == ExamDeliveryCenterBased && a.VenueID == nil && a.VenueName == "" {
		return errors.New("venue is required for centre-based exams")
	}
	if a.AllocatedAt.IsZero() {
		a.AllocatedAt = time.Now().UTC()
	}
	return nil
}

type ExamInvigilator struct {
	ID            uint      `gorm:"primaryKey"`
	ExamID        uint      `gorm:"not null;uniqueIndex:idx_exam_invigilator"`
	InvigilatorID uint      `gorm:"not null;uniqueIndex:idx_exam_invigilator"`
	AssignedBy    uint      `gorm:"not null;index"`
	AssignedAt    time.Time `gorm:"not null"`
	CreatedAt     time.Time

	Exam        Exam `gorm:"foreignKey:ExamID;constraint:OnDelete:CASCADE"`
	Invigilator User `gorm:"foreignKey:InvigilatorID;constraint:OnDelete:CASCADE"`
}

func (i *ExamInvigilator) BeforeSave(_ *gorm.DB) error {
	if i.ExamID == 0 || i.InvigilatorID == 0 || i.AssignedBy == 0 {
		return errors.New("exam_id, invigilator_id, and assigned_by are required")
	}
	if i.AssignedAt.IsZero() {
		i.AssignedAt = time.Now().UTC()
	}
	return nil
}

type ExamAttempt struct {
	ID                     uint              `gorm:"primaryKey"`
	UUID                   string            `gorm:"size:36;uniqueIndex;not null"`
	ExamID                 uint              `gorm:"not null;uniqueIndex:idx_exam_attempt_student"`
	StudentID              uint              `gorm:"not null;uniqueIndex:idx_exam_attempt_student"`
	Status                 ExamAttemptStatus `gorm:"size:40;not null;default:in_progress"`
	QuestionPayload        []byte            `gorm:"type:bytea"`
	AnswerPayload          []byte            `gorm:"type:bytea"`
	IntegrityScore         int               `gorm:"not null;default:100"`
	EnvironmentConfirmedAt *time.Time
	MonitoringArmedAt      *time.Time
	SubmittedAt            *time.Time
	SubmittedToOfficerAt   *time.Time
	SharedWithLecturerAt   *time.Time
	MarkedBy               uint    `gorm:"index"`
	Score                  float64 `gorm:"not null;default:0"`
	LecturerScore          float64 `gorm:"not null;default:0"`
	ModeratedScore         float64 `gorm:"not null;default:0"`
	ModerationComment      string  `gorm:"type:text"`
	ModeratedBy            uint    `gorm:"index"`
	ModeratedAt            *time.Time
	Feedback               string `gorm:"type:text"`
	MarkedAt               *time.Time
	TerminationReason      string `gorm:"type:text"`
	CreatedAt              time.Time
	UpdatedAt              time.Time

	Exam        Exam                   `gorm:"foreignKey:ExamID;constraint:OnDelete:CASCADE"`
	Student     User                   `gorm:"foreignKey:StudentID;constraint:OnDelete:CASCADE"`
	Alerts      []ProctoringAlert      `gorm:"foreignKey:AttemptID;constraint:OnDelete:CASCADE"`
	Annotations []ExamScriptAnnotation `gorm:"foreignKey:AttemptID;constraint:OnDelete:CASCADE"`
}

func (a *ExamAttempt) BeforeCreate(_ *gorm.DB) error {
	if strings.TrimSpace(a.UUID) == "" {
		a.UUID = uuid.NewString()
	}
	return nil
}

func (a *ExamAttempt) BeforeSave(_ *gorm.DB) error {
	a.Feedback = strings.TrimSpace(a.Feedback)
	a.ModerationComment = strings.TrimSpace(a.ModerationComment)
	a.TerminationReason = strings.TrimSpace(a.TerminationReason)
	if a.ExamID == 0 || a.StudentID == 0 {
		return errors.New("exam_id and student_id are required")
	}
	if a.Status == "" {
		a.Status = ExamAttemptInProgress
	}
	if !a.Status.Valid() {
		return errors.New("invalid exam attempt status")
	}
	return nil
}

type ExamScriptAnnotation struct {
	ID             uint    `gorm:"primaryKey"`
	UUID           string  `gorm:"size:36;uniqueIndex;not null"`
	AttemptID      uint    `gorm:"not null;index"`
	LecturerID     uint    `gorm:"not null;index"`
	QuestionID     string  `gorm:"size:120;not null;index"`
	Comment        string  `gorm:"type:text"`
	HighlightColor string  `gorm:"size:40"`
	InkPayload     []byte  `gorm:"type:bytea"`
	Score          float64 `gorm:"not null;default:0"`
	CreatedAt      time.Time
	UpdatedAt      time.Time

	Attempt  ExamAttempt `gorm:"foreignKey:AttemptID;constraint:OnDelete:CASCADE"`
	Lecturer User        `gorm:"foreignKey:LecturerID;constraint:OnDelete:CASCADE"`
}

func (a *ExamScriptAnnotation) BeforeCreate(_ *gorm.DB) error {
	if strings.TrimSpace(a.UUID) == "" {
		a.UUID = uuid.NewString()
	}
	return nil
}

func (a *ExamScriptAnnotation) BeforeSave(_ *gorm.DB) error {
	a.QuestionID = strings.TrimSpace(a.QuestionID)
	a.Comment = strings.TrimSpace(a.Comment)
	a.HighlightColor = strings.TrimSpace(a.HighlightColor)
	if a.AttemptID == 0 || a.LecturerID == 0 || a.QuestionID == "" {
		return errors.New("attempt_id, lecturer_id, and question_id are required")
	}
	return nil
}

type ProctoringAlert struct {
	ID             uint                    `gorm:"primaryKey"`
	UUID           string                  `gorm:"size:36;uniqueIndex;not null"`
	ExamID         uint                    `gorm:"not null;index"`
	AttemptID      uint                    `gorm:"not null;index"`
	StudentID      uint                    `gorm:"not null;index"`
	InvigilatorID  *uint                   `gorm:"index"`
	Severity       ProctoringAlertSeverity `gorm:"size:20;not null;default:warning"`
	EventType      string                  `gorm:"size:80;not null"`
	Message        string                  `gorm:"type:text;not null"`
	IntegrityScore int                     `gorm:"not null;default:100"`
	Evidence       map[string]any          `gorm:"serializer:json"`
	AcknowledgedBy *uint                   `gorm:"index"`
	AcknowledgedAt *time.Time
	CreatedAt      time.Time

	Exam    Exam        `gorm:"foreignKey:ExamID;constraint:OnDelete:CASCADE"`
	Attempt ExamAttempt `gorm:"foreignKey:AttemptID;constraint:OnDelete:CASCADE"`
	Student User        `gorm:"foreignKey:StudentID;constraint:OnDelete:CASCADE"`
}

func (a *ProctoringAlert) BeforeCreate(_ *gorm.DB) error {
	if strings.TrimSpace(a.UUID) == "" {
		a.UUID = uuid.NewString()
	}
	return nil
}

func (a *ProctoringAlert) BeforeSave(_ *gorm.DB) error {
	a.EventType = strings.TrimSpace(a.EventType)
	a.Message = strings.TrimSpace(a.Message)
	if a.ExamID == 0 || a.AttemptID == 0 || a.StudentID == 0 || a.EventType == "" || a.Message == "" {
		return errors.New("exam_id, attempt_id, student_id, event_type, and message are required")
	}
	if a.Severity == "" {
		a.Severity = ProctoringAlertWarning
	}
	if !a.Severity.Valid() {
		return errors.New("invalid proctoring alert severity")
	}
	return nil
}

type Result struct {
	ID                    uint                `gorm:"primaryKey"`
	UUID                  string              `gorm:"size:36;uniqueIndex;not null"`
	CourseID              uint                `gorm:"not null;index;uniqueIndex:idx_result_reference"`
	StudentID             uint                `gorm:"not null;index;uniqueIndex:idx_result_reference"`
	AssessmentType        ResultReferenceType `gorm:"size:30;not null;uniqueIndex:idx_result_reference"`
	ReferenceID           uint                `gorm:"not null;default:0;uniqueIndex:idx_result_reference"`
	Score                 float64             `gorm:"not null"`
	GradedAssessmentScore float64             `gorm:"not null;default:0"`
	AssignmentScore       float64             `gorm:"not null;default:0"`
	GroupAssignmentScore  float64             `gorm:"not null;default:0"`
	PeerReviewScore       float64             `gorm:"not null;default:0"`
	ExaminationScore      float64             `gorm:"not null;default:0"`
	TotalScore            float64             `gorm:"not null;default:0"`
	Grade                 string              `gorm:"size:10"`
	Remark                string              `gorm:"size:120"`
	Status                ResultStatus        `gorm:"size:30;not null;default:submitted;index"`
	MarkedBy              uint                `gorm:"not null;index"`
	ApprovedBy            *uint               `gorm:"index"`
	ApprovedAt            *time.Time
	PublishedAt           *time.Time
	CreatedAt             time.Time
	UpdatedAt             time.Time

	Course Course `gorm:"foreignKey:CourseID;constraint:OnDelete:CASCADE"`
}

func (r *Result) BeforeCreate(_ *gorm.DB) error {
	if strings.TrimSpace(r.UUID) == "" {
		r.UUID = uuid.NewString()
	}

	return nil
}

func (r *Result) BeforeSave(_ *gorm.DB) error {
	r.Grade = strings.ToUpper(strings.TrimSpace(r.Grade))
	r.Remark = strings.TrimSpace(r.Remark)
	if r.Status == "" {
		r.Status = ResultStatusSubmitted
	}
	if r.TotalScore == 0 {
		r.TotalScore = r.GradedAssessmentScore + r.AssignmentScore + r.GroupAssignmentScore + r.PeerReviewScore + r.ExaminationScore
	}
	if r.TotalScore == 0 {
		r.TotalScore = r.Score
	}
	if r.Score == 0 {
		r.Score = r.TotalScore
	}
	if r.Grade == "" {
		r.Grade = gradeForTotalScore(r.TotalScore)
	}
	if r.Remark == "" {
		if r.TotalScore >= 40 {
			r.Remark = "Passed"
		} else {
			r.Remark = "Failed"
		}
	}

	if r.CourseID == 0 || r.StudentID == 0 || r.MarkedBy == 0 {
		return errors.New("course_id, student_id, and marked_by are required")
	}

	if !r.AssessmentType.Valid() {
		return errors.New("invalid assessment_type")
	}
	if !r.Status.Valid() {
		return errors.New("invalid result status")
	}

	return nil
}

func gradeForTotalScore(score float64) string {
	switch {
	case score >= 70:
		return "A"
	case score >= 60:
		return "B"
	case score >= 50:
		return "C"
	case score >= 45:
		return "D"
	case score >= 40:
		return "E"
	default:
		return "F"
	}
}
