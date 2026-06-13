package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"kslasbackend/internal/database/models"
	"kslasbackend/internal/dto"
	"kslasbackend/internal/rbac"
	"kslasbackend/internal/repository"
)

type TeachingContentService struct {
	repo              *repository.TeachingRepository
	permissionService *PermissionService
}

func NewTeachingContentService(repo *repository.TeachingRepository, permissionService *PermissionService) *TeachingContentService {
	return &TeachingContentService{repo: repo, permissionService: permissionService}
}

func (s *TeachingContentService) ListVideoLectures(ctx context.Context, userID uint, filter repository.VideoLectureListFilter) ([]dto.VideoLectureResponse, error) {
	if err := s.ensureCourseRead(ctx, userID, filter.CourseID, "material.view"); err != nil {
		return nil, err
	}
	items, err := s.repo.ListVideoLectures(ctx, filter)
	if err != nil {
		return nil, err
	}
	return mapVideoLectures(items), nil
}

func (s *TeachingContentService) CreateVideoLecture(ctx context.Context, userID uint, req dto.VideoLectureCreateRequest) (*dto.VideoLectureResponse, error) {
	courseID, err := s.resolveCourseID(ctx, req.CourseID, req.CourseCode)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "course.manage_content", scopePtr(rbac.CourseScope(courseID))); err != nil {
		return nil, err
	}
	if err := s.ensureCourseLecturer(ctx, userID, courseID); err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.Title) == "" {
		return nil, ValidationError{Message: "title is required"}
	}
	if strings.TrimSpace(req.ExternalURL) == "" {
		return nil, ValidationError{Message: "external_url is required until video storage is connected"}
	}
	allowDownload := true
	if req.AllowDownload != nil {
		allowDownload = *req.AllowDownload
	}
	requireWatchedMark := true
	if req.RequireWatchedMark != nil {
		requireWatchedMark = *req.RequireWatchedMark
	}
	item := &models.VideoLecture{
		CourseID:           courseID,
		Title:              req.Title,
		Subtitle:           req.Subtitle,
		Description:        req.Description,
		LecturerName:       req.LecturerName,
		SourceType:         models.VideoSourceExternal,
		ExternalURL:        strings.TrimSpace(req.ExternalURL),
		DurationMinutes:    req.DurationMinutes,
		AudienceKeys:       req.AudienceKeys,
		Tags:               req.Tags,
		AllowDownload:      allowDownload,
		RequireWatchedMark: requireWatchedMark,
		UploadedBy:         userID,
	}
	if err := s.repo.CreateVideoLecture(ctx, item); err != nil {
		return nil, err
	}
	item, err = s.repo.GetVideoLecture(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	response := mapVideoLecture(item)
	return &response, nil
}

func (s *TeachingContentService) DeleteVideoLecture(ctx context.Context, userID, lectureID uint) error {
	item, err := s.repo.GetVideoLecture(ctx, lectureID)
	if err != nil {
		return err
	}
	if err := s.ensurePermission(ctx, userID, "course.manage_content", scopePtr(rbac.CourseScope(item.CourseID))); err != nil {
		return err
	}
	if err := s.ensureCourseLecturer(ctx, userID, item.CourseID); err != nil {
		return err
	}
	return s.repo.DeleteVideoLecture(ctx, lectureID)
}

func (s *TeachingContentService) MarkVideoLectureWatched(ctx context.Context, userID, lectureID uint, req dto.VideoLectureWatchRequest) (*dto.VideoLectureResponse, error) {
	item, err := s.repo.GetVideoLecture(ctx, lectureID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureCourseRead(ctx, userID, &item.CourseID, "material.view"); err != nil {
		return nil, err
	}
	watchedAt := time.Now().UTC()
	if req.WatchedAt != nil {
		watchedAt = *req.WatchedAt
	}
	if _, err := s.repo.UpsertVideoLectureWatch(ctx, lectureID, userID, watchedAt); err != nil {
		return nil, err
	}
	item, err = s.repo.GetVideoLecture(ctx, lectureID)
	if err != nil {
		return nil, err
	}
	response := mapVideoLecture(item)
	return &response, nil
}

func (s *TeachingContentService) ListLiveSessions(ctx context.Context, userID uint, filter repository.LiveSessionListFilter) ([]dto.LiveSessionResponse, error) {
	if err := s.ensureCourseRead(ctx, userID, filter.CourseID, "course.view"); err != nil {
		return nil, err
	}
	items, err := s.repo.ListLiveSessions(ctx, filter)
	if err != nil {
		return nil, err
	}
	return mapLiveSessions(items), nil
}

func (s *TeachingContentService) CreateLiveSession(ctx context.Context, userID uint, req dto.LiveSessionCreateRequest) (*dto.LiveSessionResponse, error) {
	if req.CourseID == 0 {
		return nil, ValidationError{Message: "course_id is required"}
	}
	if err := s.ensurePermission(ctx, userID, "liveclass.create", scopePtr(rbac.CourseScope(req.CourseID))); err != nil {
		return nil, err
	}
	if err := s.ensureCourseLecturer(ctx, userID, req.CourseID); err != nil {
		return nil, err
	}
	status := liveSessionStatusOrDefault(req.Status)
	item := &models.LiveSession{
		CourseID:     req.CourseID,
		Title:        req.Title,
		Description:  req.Description,
		LecturerName: req.LecturerName,
		RoomName:     req.RoomName,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		Status:       status,
		Agenda:       req.Agenda,
		Materials:    req.Materials,
		Settings:     mapLiveSessionSettings(req.Settings),
		CreatedBy:    userID,
	}
	if err := s.repo.CreateLiveSession(ctx, item); err != nil {
		return nil, err
	}
	item, err := s.repo.GetLiveSession(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	response := mapLiveSession(item)
	return &response, nil
}

func (s *TeachingContentService) UpdateLiveSession(ctx context.Context, userID, sessionID uint, req dto.LiveSessionUpdateRequest) (*dto.LiveSessionResponse, error) {
	item, err := s.repo.GetLiveSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if err := s.ensurePermission(ctx, userID, "liveclass.manage", scopePtr(rbac.CourseScope(item.CourseID))); err != nil {
		return nil, err
	}
	if err := s.ensureCourseLecturer(ctx, userID, item.CourseID); err != nil {
		return nil, err
	}
	item.Title = req.Title
	item.Description = req.Description
	item.LecturerName = req.LecturerName
	item.RoomName = req.RoomName
	item.StartTime = req.StartTime
	item.EndTime = req.EndTime
	item.Status = liveSessionStatusOrDefault(req.Status)
	item.Agenda = req.Agenda
	item.Materials = req.Materials
	item.Settings = mapLiveSessionSettings(req.Settings)
	if err := s.repo.UpdateLiveSession(ctx, item); err != nil {
		return nil, err
	}
	response := mapLiveSession(item)
	return &response, nil
}

func (s *TeachingContentService) ensureCourseRead(ctx context.Context, userID uint, courseID *uint, permission string) error {
	target := scopePtr(rbac.SchoolScope())
	if courseID != nil {
		target = scopePtr(rbac.CourseScope(*courseID))
	}
	if err := s.ensurePermission(ctx, userID, permission, target); err != nil {
		return err
	}
	if courseID == nil || *courseID == 0 {
		return nil
	}
	isStudent, err := s.repo.UserHasRole(ctx, userID, "student")
	if err != nil {
		return err
	}
	if !isStudent {
		return nil
	}
	registered, err := s.repo.StudentRegisteredForCourse(ctx, userID, *courseID)
	if err != nil {
		return err
	}
	if !registered {
		return ErrPermissionDenied
	}
	return nil
}

func (s *TeachingContentService) ensureCourseLecturer(ctx context.Context, userID, courseID uint) error {
	isLecturer, err := s.repo.UserHasRole(ctx, userID, "lecturer")
	if err != nil {
		return err
	}
	if !isLecturer {
		return ErrPermissionDenied
	}
	assigned, err := s.repo.LecturerAssignedToCourse(ctx, userID, courseID)
	if err != nil {
		return err
	}
	if !assigned {
		return ErrPermissionDenied
	}
	return nil
}

func (s *TeachingContentService) ensurePermission(ctx context.Context, userID uint, permissionCode string, target *rbac.Scope) error {
	allowed, err := s.permissionService.UserHasPermission(ctx, userID, permissionCode, target)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrPermissionDenied
	}
	return nil
}

func (s *TeachingContentService) resolveCourseID(ctx context.Context, courseID uint, courseCode string) (uint, error) {
	if courseID != 0 {
		return courseID, nil
	}
	code := strings.TrimSpace(courseCode)
	if code == "" {
		return 0, ValidationError{Message: "course_id or course_code is required"}
	}
	course, err := s.repo.GetCourseByCode(ctx, code)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, ValidationError{Message: "course not found"}
		}
		return 0, err
	}
	return course.ID, nil
}

func liveSessionStatusOrDefault(raw string) models.LiveSessionStatus {
	status := models.LiveSessionStatus(strings.ToLower(strings.TrimSpace(raw)))
	if status == "" {
		return models.LiveSessionStatusScheduled
	}
	if !status.Valid() {
		return models.LiveSessionStatusScheduled
	}
	return status
}

func mapLiveSessionSettings(value dto.LiveSessionSettingsPayload) models.LiveSessionSettings {
	return models.LiveSessionSettings{
		StudentCameraRequired:  value.StudentCameraRequired,
		CaptureRegistrationNo:  value.CaptureRegistrationNo,
		AllowStudentRecording:  value.AllowStudentRecording,
		AllowLecturerRecording: value.AllowLecturerRecording,
		AttendanceEnabled:      value.AttendanceEnabled,
		ChatEnabled:            value.ChatEnabled,
		QuestionsEnabled:       value.QuestionsEnabled,
	}
}

func mapLiveSessionSettingsResponse(value models.LiveSessionSettings) dto.LiveSessionSettingsPayload {
	return dto.LiveSessionSettingsPayload{
		StudentCameraRequired:  value.StudentCameraRequired,
		CaptureRegistrationNo:  value.CaptureRegistrationNo,
		AllowStudentRecording:  value.AllowStudentRecording,
		AllowLecturerRecording: value.AllowLecturerRecording,
		AttendanceEnabled:      value.AttendanceEnabled,
		ChatEnabled:            value.ChatEnabled,
		QuestionsEnabled:       value.QuestionsEnabled,
	}
}

func mapVideoLectures(items []models.VideoLecture) []dto.VideoLectureResponse {
	out := make([]dto.VideoLectureResponse, 0, len(items))
	for i := range items {
		out = append(out, mapVideoLecture(&items[i]))
	}
	return out
}

func mapVideoLecture(item *models.VideoLecture) dto.VideoLectureResponse {
	watchedBy := make([]uint, 0, len(item.Watches))
	for _, watch := range item.Watches {
		watchedBy = append(watchedBy, watch.UserID)
	}
	return dto.VideoLectureResponse{
		ID:                 item.ID,
		UUID:               item.UUID,
		CourseID:           item.CourseID,
		CourseCode:         item.Course.Code,
		CourseTitle:        item.Course.Title,
		Title:              item.Title,
		Subtitle:           item.Subtitle,
		Description:        item.Description,
		LecturerName:       item.LecturerName,
		SourceType:         string(item.SourceType),
		OriginalFileName:   item.OriginalFileName,
		ExternalURL:        item.ExternalURL,
		MimeType:           item.MimeType,
		SizeBytes:          item.SizeBytes,
		DurationMinutes:    item.DurationMinutes,
		AudienceKeys:       item.AudienceKeys,
		Tags:               item.Tags,
		AllowDownload:      item.AllowDownload,
		RequireWatchedMark: item.RequireWatchedMark,
		UploadedBy:         item.UploadedBy,
		PublishedAt:        item.PublishedAt,
		StreamURL:          item.ExternalURL,
		WatchedCount:       len(item.Watches),
		WatchedByUserIDs:   watchedBy,
		CreatedAt:          item.CreatedAt,
		UpdatedAt:          item.UpdatedAt,
	}
}

func mapLiveSessions(items []models.LiveSession) []dto.LiveSessionResponse {
	out := make([]dto.LiveSessionResponse, 0, len(items))
	for i := range items {
		out = append(out, mapLiveSession(&items[i]))
	}
	return out
}

func mapLiveSession(item *models.LiveSession) dto.LiveSessionResponse {
	return dto.LiveSessionResponse{
		ID:           item.ID,
		UUID:         item.UUID,
		CourseID:     item.CourseID,
		CourseCode:   item.Course.Code,
		CourseTitle:  item.Course.Title,
		Title:        item.Title,
		Description:  item.Description,
		LecturerName: item.LecturerName,
		RoomName:     item.RoomName,
		StartTime:    item.StartTime,
		EndTime:      item.EndTime,
		Status:       string(item.Status),
		Agenda:       item.Agenda,
		Materials:    item.Materials,
		Settings:     mapLiveSessionSettingsResponse(item.Settings),
		CreatedBy:    item.CreatedBy,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
	}
}
