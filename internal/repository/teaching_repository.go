package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"kslasbackend/internal/database/models"
)

type TeachingRepository struct {
	db *gorm.DB
}

type MaterialListFilter struct {
	CourseID      *uint
	CourseCode    string
	PublishedOnly bool
}

type VideoLectureListFilter struct {
	CourseID *uint
}

type LiveSessionListFilter struct {
	CourseID *uint
	Status   string
}

type AssignmentListFilter struct {
	CourseID *uint
	Status   string
}

type QuizListFilter struct {
	CourseID *uint
	Status   string
}

type ExamListFilter struct {
	CourseID *uint
	Status   string
}

type ResultListFilter struct {
	CourseID       *uint
	StudentID      *uint
	AssessmentType string
	Status         string
}

type ResultStats struct {
	Count        int64
	AverageScore *float64
}

type LecturerCourseAssignment struct {
	CourseID    uint
	CourseCode  string
	CourseTitle string
	LecturerID  uint
	FirstName   string
	LastName    string
	Email       string
}

type LecturerCourseMetrics struct {
	AssignmentsPublished      int64
	AssignmentSubmissionCount int64
	AssignmentsMarked         int64
	ForumDiscussionCount      int64
	ExamCount                 int64
	ExamMarkedCount           int64
	AverageExamScore          *float64
	LiveSessionsScheduled     int64
	LiveSessionsConducted     int64
	VideoLecturesUploaded     int64
	MaterialsUploaded         int64
	LastAssignmentAt          *time.Time
	LastMarkedAt              *time.Time
	LastForumAt               *time.Time
	LastExamAt                *time.Time
	LastLiveSessionAt         *time.Time
	LastVideoLectureAt        *time.Time
	LastMaterialAt            *time.Time
}

func NewTeachingRepository(db *gorm.DB) *TeachingRepository {
	return &TeachingRepository{db: db}
}

func (r *TeachingRepository) GetUser(ctx context.Context, userID uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Select("id", "first_name", "last_name").
		First(&user, userID).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *TeachingRepository) GetCourse(ctx context.Context, courseID uint) (*models.Course, error) {
	var course models.Course
	err := r.db.WithContext(ctx).First(&course, courseID).Error
	if err != nil {
		return nil, err
	}

	return &course, nil
}

func (r *TeachingRepository) GetCourseByCode(ctx context.Context, code string) (*models.Course, error) {
	var course models.Course
	err := r.db.WithContext(ctx).
		Where("LOWER(code) = LOWER(?)", code).
		First(&course).Error
	if err != nil {
		return nil, err
	}

	return &course, nil
}

func (r *TeachingRepository) GetRoleByCode(ctx context.Context, code string) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).
		Where("LOWER(code) = LOWER(?)", strings.TrimSpace(code)).
		First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *TeachingRepository) CreateUserWithRole(ctx context.Context, user *models.User, roleCode string, scopeType models.ScopeType, scopeID *uint, assignedBy uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		var role models.Role
		if err := tx.Where("LOWER(code) = LOWER(?)", strings.TrimSpace(roleCode)).First(&role).Error; err != nil {
			return err
		}
		return tx.Create(&models.UserRole{
			UserID:     user.ID,
			RoleID:     role.ID,
			ScopeType:  scopeType,
			ScopeID:    scopeID,
			IsPrimary:  true,
			AssignedBy: &assignedBy,
		}).Error
	})
}

func (r *TeachingRepository) CreateStudentWithProfile(ctx context.Context, user *models.User, profile *models.StudentAcademicProfile, assignedBy uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		var role models.Role
		if err := tx.Where("code = ?", "student").First(&role).Error; err != nil {
			return err
		}
		if err := tx.Create(&models.UserRole{
			UserID:     user.ID,
			RoleID:     role.ID,
			ScopeType:  models.ScopeProgramme,
			ScopeID:    &profile.ProgrammeID,
			IsPrimary:  true,
			AssignedBy: &assignedBy,
		}).Error; err != nil {
			return err
		}
		profile.StudentID = user.ID
		return tx.Create(profile).Error
	})
}

func (r *TeachingRepository) AssignLecturerToCourse(ctx context.Context, courseID, lecturerID, assignedBy uint) error {
	return r.db.WithContext(ctx).Create(&models.CourseLecturerAssignment{
		CourseID:   courseID,
		LecturerID: lecturerID,
		AssignedBy: assignedBy,
	}).Error
}

func (r *TeachingRepository) GetStudentAcademicProfile(ctx context.Context, studentID uint) (*models.StudentAcademicProfile, error) {
	var profile models.StudentAcademicProfile
	err := r.db.WithContext(ctx).
		Preload("Programme").
		Where("student_id = ?", studentID).
		First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *TeachingRepository) ListEligibleCoursesForStudent(ctx context.Context, profile models.StudentAcademicProfile, semester string) ([]models.Course, error) {
	query := r.db.WithContext(ctx).Model(&models.Course{}).
		Where("is_active = ?", true).
		Where("level = ?", profile.Level).
		Where("(programme_id = ? OR programme_id IS NULL)", profile.ProgrammeID)
	if strings.TrimSpace(semester) != "" {
		query = query.Where("semester = ?", strings.TrimSpace(semester))
	} else {
		query = query.Where("semester = ?", profile.Semester)
	}
	var items []models.Course
	err := query.Order("code ASC").Find(&items).Error
	return items, err
}

func (r *TeachingRepository) RegisterStudentCourses(ctx context.Context, profile models.StudentAcademicProfile, courseIDs []uint, academicSession, semester string) ([]models.CourseRegistration, error) {
	registrations := make([]models.CourseRegistration, 0, len(courseIDs))
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, courseID := range courseIDs {
			var course models.Course
			if err := tx.Where("id = ? AND is_active = ? AND level = ? AND (programme_id = ? OR programme_id IS NULL)", courseID, true, profile.Level, profile.ProgrammeID).First(&course).Error; err != nil {
				return err
			}
			registration := models.CourseRegistration{
				StudentID:       profile.StudentID,
				CourseID:        courseID,
				AcademicSession: academicSession,
				Semester:        semester,
				Level:           profile.Level,
				Status:          models.CourseRegistrationPending,
				RegisteredAt:    time.Now().UTC(),
			}
			if err := tx.Where("student_id = ? AND course_id = ? AND academic_session = ?", profile.StudentID, courseID, academicSession).
				Assign(registration).
				FirstOrCreate(&registration).Error; err != nil {
				return err
			}
			registration.Course = course
			registrations = append(registrations, registration)
		}
		return nil
	})
	return registrations, err
}

func (r *TeachingRepository) ListStudentCourseRegistrations(ctx context.Context, studentID uint) ([]models.CourseRegistration, error) {
	var items []models.CourseRegistration
	err := r.db.WithContext(ctx).
		Preload("Course").
		Where("student_id = ?", studentID).
		Order("created_at DESC").
		Find(&items).Error
	return items, err
}

func (r *TeachingRepository) ListCourseRegistrations(ctx context.Context, courseID uint) ([]models.CourseRegistration, error) {
	var items []models.CourseRegistration
	err := r.db.WithContext(ctx).
		Preload("Student").
		Preload("Course").
		Where("course_id = ?", courseID).
		Where("status IN ?", []models.CourseRegistrationStatus{
			models.CourseRegistrationPending,
			models.CourseRegistrationApproved,
		}).
		Order("created_at DESC").
		Find(&items).Error
	return items, err
}

func (r *TeachingRepository) UserHasRole(ctx context.Context, userID uint, roleCode string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("user_roles").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND roles.code = ?", userID, strings.TrimSpace(roleCode)).
		Count(&count).Error
	return count > 0, err
}

func (r *TeachingRepository) LecturerAssignedToCourse(ctx context.Context, lecturerID, courseID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.CourseLecturerAssignment{}).
		Where("lecturer_id = ? AND course_id = ?", lecturerID, courseID).
		Count(&count).Error
	return count > 0, err
}

func (r *TeachingRepository) StudentRegisteredForCourse(ctx context.Context, studentID, courseID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.CourseRegistration{}).
		Where("student_id = ? AND course_id = ?", studentID, courseID).
		Where("status IN ?", []models.CourseRegistrationStatus{
			models.CourseRegistrationPending,
			models.CourseRegistrationApproved,
		}).
		Count(&count).Error
	return count > 0, err
}

func (r *TeachingRepository) ListMaterials(ctx context.Context, filter MaterialListFilter) ([]models.CourseMaterial, error) {
	query := r.db.WithContext(ctx).Model(&models.CourseMaterial{}).Preload("Course")
	if filter.CourseID != nil {
		query = query.Where("course_id = ?", *filter.CourseID)
	}
	if filter.PublishedOnly {
		query = query.Where("published_at IS NOT NULL")
	}

	var items []models.CourseMaterial
	err := query.Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *TeachingRepository) GetMaterial(ctx context.Context, materialID uint) (*models.CourseMaterial, error) {
	var item models.CourseMaterial
	err := r.db.WithContext(ctx).
		Preload("Course").
		First(&item, materialID).Error
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *TeachingRepository) CreateMaterial(ctx context.Context, item *models.CourseMaterial) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *TeachingRepository) UpdateMaterial(ctx context.Context, item *models.CourseMaterial) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *TeachingRepository) DeleteMaterial(ctx context.Context, materialID uint) error {
	return r.db.WithContext(ctx).Delete(&models.CourseMaterial{}, materialID).Error
}

func (r *TeachingRepository) ListVideoLectures(ctx context.Context, filter VideoLectureListFilter) ([]models.VideoLecture, error) {
	query := r.db.WithContext(ctx).Model(&models.VideoLecture{}).Preload("Course").Preload("Watches")
	if filter.CourseID != nil {
		query = query.Where("course_id = ?", *filter.CourseID)
	}

	var items []models.VideoLecture
	err := query.Order("published_at DESC").Find(&items).Error
	return items, err
}

func (r *TeachingRepository) GetVideoLecture(ctx context.Context, lectureID uint) (*models.VideoLecture, error) {
	var item models.VideoLecture
	err := r.db.WithContext(ctx).
		Preload("Course").
		Preload("Watches").
		First(&item, lectureID).Error
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *TeachingRepository) CreateVideoLecture(ctx context.Context, item *models.VideoLecture) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *TeachingRepository) DeleteVideoLecture(ctx context.Context, lectureID uint) error {
	return r.db.WithContext(ctx).Delete(&models.VideoLecture{}, lectureID).Error
}

func (r *TeachingRepository) UpsertVideoLectureWatch(ctx context.Context, lectureID, userID uint, watchedAt time.Time) (*models.VideoLectureWatch, error) {
	tx := r.db.WithContext(ctx)

	var item models.VideoLectureWatch
	err := tx.Where("video_lecture_id = ? AND user_id = ?", lectureID, userID).First(&item).Error
	switch {
	case err == nil:
		item.WatchedAt = watchedAt
		if err := tx.Save(&item).Error; err != nil {
			return nil, err
		}
	case err == gorm.ErrRecordNotFound:
		item = models.VideoLectureWatch{
			VideoLectureID: lectureID,
			UserID:         userID,
			WatchedAt:      watchedAt,
		}
		if err := tx.Create(&item).Error; err != nil {
			return nil, err
		}
	default:
		return nil, err
	}

	return &item, nil
}

func (r *TeachingRepository) ListLiveSessions(ctx context.Context, filter LiveSessionListFilter) ([]models.LiveSession, error) {
	query := r.db.WithContext(ctx).Model(&models.LiveSession{}).Preload("Course")
	if filter.CourseID != nil {
		query = query.Where("course_id = ?", *filter.CourseID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	var items []models.LiveSession
	err := query.Order("start_time DESC").Find(&items).Error
	return items, err
}

func (r *TeachingRepository) GetLiveSession(ctx context.Context, sessionID uint) (*models.LiveSession, error) {
	var item models.LiveSession
	err := r.db.WithContext(ctx).
		Preload("Course").
		Preload("Attendance").
		Preload("Recordings").
		First(&item, sessionID).Error
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *TeachingRepository) CreateLiveSession(ctx context.Context, item *models.LiveSession) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *TeachingRepository) UpdateLiveSession(ctx context.Context, item *models.LiveSession) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *TeachingRepository) DeleteLiveSession(ctx context.Context, sessionID uint) error {
	return r.db.WithContext(ctx).Delete(&models.LiveSession{}, sessionID).Error
}

func (r *TeachingRepository) ListAttendance(ctx context.Context, sessionID uint) ([]models.LiveSessionAttendance, error) {
	var items []models.LiveSessionAttendance
	err := r.db.WithContext(ctx).
		Where("live_session_id = ?", sessionID).
		Order("updated_at DESC").
		Find(&items).Error
	return items, err
}

func (r *TeachingRepository) UpsertAttendanceRecords(ctx context.Context, sessionID, capturedBy uint, items []models.LiveSessionAttendance) ([]models.LiveSessionAttendance, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	out := make([]models.LiveSessionAttendance, 0, len(items))
	for _, item := range items {
		var existing models.LiveSessionAttendance
		err := tx.Where("live_session_id = ? AND user_id = ?", sessionID, item.UserID).First(&existing).Error
		switch {
		case err == nil:
			existing.RegistrationNumber = item.RegistrationNumber
			existing.Status = item.Status
			existing.JoinedAt = item.JoinedAt
			existing.LeftAt = item.LeftAt
			existing.DurationMinutes = item.DurationMinutes
			existing.CapturedBy = capturedBy
			if err := tx.Save(&existing).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
			out = append(out, existing)
		case err == gorm.ErrRecordNotFound:
			item.LiveSessionID = sessionID
			item.CapturedBy = capturedBy
			if err := tx.Create(&item).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
			out = append(out, item)
		default:
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return out, nil
}

func (r *TeachingRepository) CreateRecording(ctx context.Context, item *models.LiveSessionRecording) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *TeachingRepository) ListRecordings(ctx context.Context, sessionID uint) ([]models.LiveSessionRecording, error) {
	var items []models.LiveSessionRecording
	err := r.db.WithContext(ctx).
		Where("live_session_id = ?", sessionID).
		Order("published_at DESC").
		Find(&items).Error
	return items, err
}

func (r *TeachingRepository) GetRecording(ctx context.Context, recordingID uint) (*models.LiveSessionRecording, error) {
	var item models.LiveSessionRecording
	err := r.db.WithContext(ctx).First(&item, recordingID).Error
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *TeachingRepository) ListCourseForumPosts(ctx context.Context, courseID uint) ([]models.CourseForumPost, error) {
	var items []models.CourseForumPost
	err := r.db.WithContext(ctx).
		Preload("Course").
		Preload("Author.UserRoles.Role").
		Preload("Replies", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Preload("Replies.Author.UserRoles.Role").
		Where("course_id = ? AND parent_id IS NULL", courseID).
		Order("is_pinned DESC, created_at DESC").
		Find(&items).Error
	return items, err
}

func (r *TeachingRepository) GetCourseForumPost(ctx context.Context, postID uint) (*models.CourseForumPost, error) {
	var item models.CourseForumPost
	err := r.db.WithContext(ctx).
		Preload("Course").
		Preload("Author.UserRoles.Role").
		First(&item, postID).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *TeachingRepository) CreateCourseForumPost(ctx context.Context, item *models.CourseForumPost) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *TeachingRepository) UpdateCourseForumPost(ctx context.Context, item *models.CourseForumPost) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *TeachingRepository) ListCourseDirectMessages(ctx context.Context, courseID, userID uint, withUserID *uint) ([]models.CourseDirectMessage, error) {
	query := r.db.WithContext(ctx).
		Model(&models.CourseDirectMessage{}).
		Preload("Course").
		Preload("Sender").
		Preload("Recipient").
		Where("course_id = ? AND (sender_id = ? OR recipient_id = ?)", courseID, userID, userID)
	if withUserID != nil {
		query = query.Where("(sender_id = ? OR recipient_id = ?)", *withUserID, *withUserID)
	}
	var items []models.CourseDirectMessage
	err := query.Order("created_at ASC").Find(&items).Error
	return items, err
}

func (r *TeachingRepository) CreateCourseDirectMessage(ctx context.Context, item *models.CourseDirectMessage) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *TeachingRepository) CountCourseForumPosts(ctx context.Context, courseID uint) (int64, error) {
	return r.countByCourse(ctx, &models.CourseForumPost{}, courseID)
}

func (r *TeachingRepository) ListAssignments(ctx context.Context, filter AssignmentListFilter) ([]models.Assignment, error) {
	query := r.db.WithContext(ctx).
		Model(&models.Assignment{}).
		Preload("Course").
		Preload("Groups.Members").
		Preload("Submissions.Files").
		Preload("PeerReviews").
		Preload("Grades")
	if filter.CourseID != nil {
		query = query.Where("course_id = ?", *filter.CourseID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	var items []models.Assignment
	err := query.Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *TeachingRepository) GetAssignment(ctx context.Context, assignmentID uint) (*models.Assignment, error) {
	var item models.Assignment
	err := r.db.WithContext(ctx).
		Preload("Course").
		Preload("Groups.Members").
		Preload("Submissions.Files").
		Preload("PeerReviews").
		Preload("Grades").
		First(&item, assignmentID).Error
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *TeachingRepository) CreateAssignment(ctx context.Context, item *models.Assignment) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *TeachingRepository) UpdateAssignment(ctx context.Context, item *models.Assignment) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *TeachingRepository) DeleteAssignment(ctx context.Context, assignmentID uint) error {
	return r.db.WithContext(ctx).Delete(&models.Assignment{}, assignmentID).Error
}

func (r *TeachingRepository) ReplaceAssignmentGroups(ctx context.Context, assignmentID, createdBy uint, groups []models.AssignmentGroup) ([]models.AssignmentGroup, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	if err := tx.Where("assignment_id = ?", assignmentID).Delete(&models.AssignmentGroup{}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	for i := range groups {
		groups[i].AssignmentID = assignmentID
		groups[i].CreatedBy = createdBy
		for j := range groups[i].Members {
			groups[i].Members[j].AddedBy = createdBy
		}
		if err := tx.Create(&groups[i]).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return groups, nil
}

func (r *TeachingRepository) ReplaceAssignmentPeerReviews(ctx context.Context, assignmentID, assignedBy uint, reviews []models.AssignmentPeerReview) ([]models.AssignmentPeerReview, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	if err := tx.Where("assignment_id = ?", assignmentID).Delete(&models.AssignmentPeerReview{}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	for i := range reviews {
		reviews[i].AssignmentID = assignmentID
		reviews[i].AssignedBy = assignedBy
		if err := tx.Create(&reviews[i]).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return reviews, nil
}

func (r *TeachingRepository) ListAssignmentSubmissions(ctx context.Context, assignmentID uint) ([]models.AssignmentSubmission, error) {
	var items []models.AssignmentSubmission
	err := r.db.WithContext(ctx).
		Preload("Files").
		Where("assignment_id = ?", assignmentID).
		Order("submitted_at DESC").
		Find(&items).Error
	return items, err
}

func (r *TeachingRepository) UpsertAssignmentSubmission(ctx context.Context, item *models.AssignmentSubmission) (*models.AssignmentSubmission, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	var existing models.AssignmentSubmission
	query := tx.Where("assignment_id = ? AND student_id = ?", item.AssignmentID, item.StudentID)
	if item.GroupID == nil {
		query = query.Where("group_id IS NULL")
	} else {
		query = query.Where("group_id = ?", *item.GroupID)
	}

	err := query.First(&existing).Error
	switch {
	case err == nil:
		existing.TextAnswer = item.TextAnswer
		existing.WhiteboardData = item.WhiteboardData
		existing.Status = models.AssignmentSubmissionStatusResubmitted
		existing.SubmittedAt = item.SubmittedAt
		if err := tx.Save(&existing).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		if err := tx.Where("submission_id = ?", existing.ID).Delete(&models.AssignmentSubmissionFile{}).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		for i := range item.Files {
			item.Files[i].SubmissionID = existing.ID
			if err := tx.Create(&item.Files[i]).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		}
		item = &existing
	case err == gorm.ErrRecordNotFound:
		if err := tx.Create(item).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	default:
		tx.Rollback()
		return nil, err
	}

	if err := tx.Preload("Files").First(item, item.ID).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return item, nil
}

func (r *TeachingRepository) GetAssignmentPeerReview(ctx context.Context, assignmentID, reviewID uint) (*models.AssignmentPeerReview, error) {
	var item models.AssignmentPeerReview
	err := r.db.WithContext(ctx).
		Where("assignment_id = ?", assignmentID).
		First(&item, reviewID).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *TeachingRepository) UpdateAssignmentPeerReview(ctx context.Context, item *models.AssignmentPeerReview) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *TeachingRepository) ListAssignmentGrades(ctx context.Context, assignmentID uint) ([]models.AssignmentGrade, error) {
	var items []models.AssignmentGrade
	err := r.db.WithContext(ctx).
		Where("assignment_id = ?", assignmentID).
		Order("updated_at DESC").
		Find(&items).Error
	return items, err
}

func (r *TeachingRepository) UpsertAssignmentGrades(ctx context.Context, assignmentID, markerID uint, items []models.AssignmentGrade) ([]models.AssignmentGrade, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	out := make([]models.AssignmentGrade, 0, len(items))
	for _, item := range items {
		var existing models.AssignmentGrade
		err := tx.Where("assignment_id = ? AND student_id = ?", assignmentID, item.StudentID).First(&existing).Error
		switch {
		case err == nil:
			existing.Score = item.Score
			existing.Feedback = item.Feedback
			existing.Status = item.Status
			existing.MarkedAt = item.MarkedAt
			existing.MarkerID = markerID
			if err := tx.Save(&existing).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
			out = append(out, existing)
		case err == gorm.ErrRecordNotFound:
			item.AssignmentID = assignmentID
			item.MarkerID = markerID
			if err := tx.Create(&item).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
			out = append(out, item)
		default:
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return out, nil
}

func (r *TeachingRepository) ListQuizzes(ctx context.Context, filter QuizListFilter) ([]models.Quiz, error) {
	query := r.db.WithContext(ctx).Model(&models.Quiz{}).Preload("Course")
	if filter.CourseID != nil {
		query = query.Where("course_id = ?", *filter.CourseID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	var items []models.Quiz
	err := query.Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *TeachingRepository) GetQuiz(ctx context.Context, quizID uint) (*models.Quiz, error) {
	var item models.Quiz
	err := r.db.WithContext(ctx).Preload("Course").First(&item, quizID).Error
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *TeachingRepository) CreateQuiz(ctx context.Context, item *models.Quiz) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *TeachingRepository) UpdateQuiz(ctx context.Context, item *models.Quiz) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *TeachingRepository) DeleteQuiz(ctx context.Context, quizID uint) error {
	return r.db.WithContext(ctx).Delete(&models.Quiz{}, quizID).Error
}

func (r *TeachingRepository) ListExamVenues(ctx context.Context, activeOnly bool) ([]models.ExamVenue, error) {
	query := r.db.WithContext(ctx).Model(&models.ExamVenue{})
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}
	var items []models.ExamVenue
	err := query.Order("country ASC, city ASC, name ASC").Find(&items).Error
	return items, err
}

func (r *TeachingRepository) GetExamVenue(ctx context.Context, venueID uint) (*models.ExamVenue, error) {
	var item models.ExamVenue
	err := r.db.WithContext(ctx).First(&item, venueID).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *TeachingRepository) CreateExamVenue(ctx context.Context, item *models.ExamVenue) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *TeachingRepository) ListExams(ctx context.Context, filter ExamListFilter) ([]models.Exam, error) {
	query := r.db.WithContext(ctx).Model(&models.Exam{}).Preload("Course").Preload("Invigilators")
	if filter.CourseID != nil {
		query = query.Where("course_id = ?", *filter.CourseID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	var items []models.Exam
	err := query.Order("start_time DESC").Find(&items).Error
	return items, err
}

func (r *TeachingRepository) GetExam(ctx context.Context, examID uint) (*models.Exam, error) {
	var item models.Exam
	err := r.db.WithContext(ctx).Preload("Course").Preload("Invigilators").First(&item, examID).Error
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *TeachingRepository) CreateExam(ctx context.Context, item *models.Exam) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *TeachingRepository) UpdateExam(ctx context.Context, item *models.Exam) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *TeachingRepository) DeleteExam(ctx context.Context, examID uint) error {
	return r.db.WithContext(ctx).Delete(&models.Exam{}, examID).Error
}

func (r *TeachingRepository) UpsertExamStudentAllocations(ctx context.Context, examID, allocatedBy uint, items []models.ExamStudentAllocation) ([]models.ExamStudentAllocation, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	out := make([]models.ExamStudentAllocation, 0, len(items))
	for _, item := range items {
		var existing models.ExamStudentAllocation
		err := tx.Where("exam_id = ? AND student_id = ?", examID, item.StudentID).First(&existing).Error
		switch {
		case err == nil:
			existing.DeliveryMode = item.DeliveryMode
			existing.VenueID = item.VenueID
			existing.VenueName = item.VenueName
			existing.VenueAddress = item.VenueAddress
			existing.IsInternational = item.IsInternational
			existing.CountryOfResidence = item.CountryOfResidence
			existing.AllocatedBy = allocatedBy
			existing.AllocatedAt = time.Now().UTC()
			if err := tx.Save(&existing).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
			out = append(out, existing)
		case err == gorm.ErrRecordNotFound:
			item.ExamID = examID
			item.AllocatedBy = allocatedBy
			if item.AllocatedAt.IsZero() {
				item.AllocatedAt = time.Now().UTC()
			}
			if err := tx.Create(&item).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
			out = append(out, item)
		default:
			tx.Rollback()
			return nil, err
		}
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return out, nil
}

func (r *TeachingRepository) ListExamStudentAllocations(ctx context.Context, examID uint) ([]models.ExamStudentAllocation, error) {
	var items []models.ExamStudentAllocation
	err := r.db.WithContext(ctx).
		Preload("Venue").
		Where("exam_id = ?", examID).
		Order("updated_at DESC").
		Find(&items).Error
	return items, err
}

func (r *TeachingRepository) ReplaceExamInvigilators(ctx context.Context, examID, assignedBy uint, invigilatorIDs []uint) ([]models.ExamInvigilator, error) {
	var out []models.ExamInvigilator
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("exam_id = ?", examID).Delete(&models.ExamInvigilator{}).Error; err != nil {
			return err
		}
		now := time.Now().UTC()
		for _, invigilatorID := range invigilatorIDs {
			if invigilatorID == 0 {
				continue
			}
			out = append(out, models.ExamInvigilator{
				ExamID:        examID,
				InvigilatorID: invigilatorID,
				AssignedBy:    assignedBy,
				AssignedAt:    now,
			})
		}
		if len(out) == 0 {
			return nil
		}
		return tx.Create(&out).Error
	})
	return out, err
}

func (r *TeachingRepository) GetExamAttempt(ctx context.Context, attemptID uint) (*models.ExamAttempt, error) {
	var item models.ExamAttempt
	err := r.db.WithContext(ctx).
		Preload("Exam").
		Preload("Exam.Course").
		Preload("Exam.Invigilators").
		First(&item, attemptID).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *TeachingRepository) GetExamAttemptForStudent(ctx context.Context, examID, studentID uint) (*models.ExamAttempt, error) {
	var item models.ExamAttempt
	err := r.db.WithContext(ctx).
		Where("exam_id = ? AND student_id = ?", examID, studentID).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *TeachingRepository) ListExamAttempts(ctx context.Context, examID uint) ([]models.ExamAttempt, error) {
	var items []models.ExamAttempt
	err := r.db.WithContext(ctx).
		Where("exam_id = ?", examID).
		Order("created_at DESC").
		Find(&items).Error
	return items, err
}

func (r *TeachingRepository) UpsertExamAttempt(ctx context.Context, item *models.ExamAttempt) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *TeachingRepository) CreateExamScriptAnnotation(ctx context.Context, item *models.ExamScriptAnnotation) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *TeachingRepository) ListExamScriptAnnotations(ctx context.Context, attemptID uint) ([]models.ExamScriptAnnotation, error) {
	var items []models.ExamScriptAnnotation
	err := r.db.WithContext(ctx).
		Where("attempt_id = ?", attemptID).
		Order("created_at ASC").
		Find(&items).Error
	return items, err
}

func (r *TeachingRepository) CreateProctoringAlert(ctx context.Context, item *models.ProctoringAlert) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *TeachingRepository) ListProctoringAlerts(ctx context.Context, invigilatorID *uint, acknowledged *bool) ([]models.ProctoringAlert, error) {
	query := r.db.WithContext(ctx).Model(&models.ProctoringAlert{})
	if invigilatorID != nil {
		query = query.Where("invigilator_id = ? OR invigilator_id IS NULL", *invigilatorID)
	}
	if acknowledged != nil {
		if *acknowledged {
			query = query.Where("acknowledged_at IS NOT NULL")
		} else {
			query = query.Where("acknowledged_at IS NULL")
		}
	}
	var items []models.ProctoringAlert
	err := query.Order("created_at DESC").Find(&items).Error
	return items, err
}

func (r *TeachingRepository) AcknowledgeProctoringAlert(ctx context.Context, alertID, userID uint) (*models.ProctoringAlert, error) {
	var item models.ProctoringAlert
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&item, alertID).Error; err != nil {
			return err
		}
		now := time.Now().UTC()
		item.AcknowledgedBy = &userID
		item.AcknowledgedAt = &now
		return tx.Save(&item).Error
	})
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *TeachingRepository) ListResults(ctx context.Context, filter ResultListFilter) ([]models.Result, error) {
	query := r.db.WithContext(ctx).Model(&models.Result{}).Preload("Course")
	if filter.CourseID != nil {
		query = query.Where("course_id = ?", *filter.CourseID)
	}
	if filter.StudentID != nil {
		query = query.Where("student_id = ?", *filter.StudentID)
	}
	if filter.AssessmentType != "" {
		query = query.Where("assessment_type = ?", filter.AssessmentType)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	var items []models.Result
	err := query.Order("updated_at DESC").Find(&items).Error
	return items, err
}

func (r *TeachingRepository) GetResult(ctx context.Context, resultID uint) (*models.Result, error) {
	var item models.Result
	err := r.db.WithContext(ctx).Preload("Course").First(&item, resultID).Error
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *TeachingRepository) CreateResult(ctx context.Context, item *models.Result) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *TeachingRepository) UpdateResult(ctx context.Context, item *models.Result) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *TeachingRepository) CountCourseMaterials(ctx context.Context, courseID uint) (int64, error) {
	return r.countByCourse(ctx, &models.CourseMaterial{}, courseID)
}

func (r *TeachingRepository) CountVideoLectures(ctx context.Context, courseID uint) (int64, error) {
	return r.countByCourse(ctx, &models.VideoLecture{}, courseID)
}

func (r *TeachingRepository) CountLiveSessions(ctx context.Context, courseID uint) (int64, error) {
	return r.countByCourse(ctx, &models.LiveSession{}, courseID)
}

func (r *TeachingRepository) CountLiveSessionRecordings(ctx context.Context, courseID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("live_session_recordings").
		Joins("JOIN live_sessions ON live_sessions.id = live_session_recordings.live_session_id").
		Where("live_sessions.course_id = ?", courseID).
		Count(&count).Error
	return count, err
}

func (r *TeachingRepository) CountAttendance(ctx context.Context, courseID uint, status *models.AttendanceStatus) (int64, error) {
	query := r.db.WithContext(ctx).
		Table("live_session_attendances").
		Joins("JOIN live_sessions ON live_sessions.id = live_session_attendances.live_session_id").
		Where("live_sessions.course_id = ?", courseID)
	if status != nil {
		query = query.Where("live_session_attendances.status = ?", *status)
	}

	var count int64
	err := query.Count(&count).Error
	return count, err
}

func (r *TeachingRepository) CountAssignments(ctx context.Context, courseID uint) (int64, error) {
	return r.countByCourse(ctx, &models.Assignment{}, courseID)
}

func (r *TeachingRepository) CountMarkedAssignmentGrades(ctx context.Context, courseID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("assignment_grades").
		Joins("JOIN assignments ON assignments.id = assignment_grades.assignment_id").
		Where("assignments.course_id = ?", courseID).
		Count(&count).Error
	return count, err
}

func (r *TeachingRepository) CountQuizzes(ctx context.Context, courseID uint) (int64, error) {
	return r.countByCourse(ctx, &models.Quiz{}, courseID)
}

func (r *TeachingRepository) CountExams(ctx context.Context, courseID uint) (int64, error) {
	return r.countByCourse(ctx, &models.Exam{}, courseID)
}

func (r *TeachingRepository) ResultMetrics(ctx context.Context, courseID uint) (ResultStats, error) {
	type stats struct {
		Count        int64
		AverageScore *float64
	}

	var row stats
	err := r.db.WithContext(ctx).
		Model(&models.Result{}).
		Select("COUNT(*) AS count, AVG(score) AS average_score").
		Where("course_id = ?", courseID).
		Scan(&row).Error
	if err != nil {
		return ResultStats{}, err
	}

	return ResultStats{
		Count:        row.Count,
		AverageScore: row.AverageScore,
	}, nil
}

func (r *TeachingRepository) LatestCourseActivity(ctx context.Context, courseID uint) (*time.Time, error) {
	latest := make([]time.Time, 0, 7)

	addIfPresent := func(query *gorm.DB) error {
		var value *time.Time
		if err := query.Scan(&value).Error; err != nil {
			return err
		}
		if value != nil {
			latest = append(latest, *value)
		}
		return nil
	}

	if err := addIfPresent(r.db.WithContext(ctx).Model(&models.CourseMaterial{}).Select("MAX(updated_at)").Where("course_id = ?", courseID)); err != nil {
		return nil, err
	}
	if err := addIfPresent(r.db.WithContext(ctx).Model(&models.VideoLecture{}).Select("MAX(updated_at)").Where("course_id = ?", courseID)); err != nil {
		return nil, err
	}
	if err := addIfPresent(r.db.WithContext(ctx).Model(&models.LiveSession{}).Select("MAX(updated_at)").Where("course_id = ?", courseID)); err != nil {
		return nil, err
	}
	if err := addIfPresent(r.db.WithContext(ctx).Model(&models.Assignment{}).Select("MAX(updated_at)").Where("course_id = ?", courseID)); err != nil {
		return nil, err
	}
	if err := addIfPresent(r.db.WithContext(ctx).Model(&models.Quiz{}).Select("MAX(updated_at)").Where("course_id = ?", courseID)); err != nil {
		return nil, err
	}
	if err := addIfPresent(r.db.WithContext(ctx).Model(&models.Exam{}).Select("MAX(updated_at)").Where("course_id = ?", courseID)); err != nil {
		return nil, err
	}
	if err := addIfPresent(r.db.WithContext(ctx).Model(&models.Result{}).Select("MAX(updated_at)").Where("course_id = ?", courseID)); err != nil {
		return nil, err
	}

	if len(latest) == 0 {
		return nil, nil
	}

	maxValue := latest[0]
	for _, item := range latest[1:] {
		if item.After(maxValue) {
			maxValue = item
		}
	}

	return &maxValue, nil
}

func (r *TeachingRepository) ListLecturerCourseAssignments(ctx context.Context, courseID, lecturerID *uint) ([]LecturerCourseAssignment, error) {
	query := r.db.WithContext(ctx).
		Table("course_lecturer_assignments AS cla").
		Select("cla.course_id, courses.code AS course_code, courses.title AS course_title, cla.lecturer_id, users.first_name, users.last_name, users.email").
		Joins("JOIN courses ON courses.id = cla.course_id").
		Joins("JOIN users ON users.id = cla.lecturer_id")
	if courseID != nil {
		query = query.Where("cla.course_id = ?", *courseID)
	}
	if lecturerID != nil {
		query = query.Where("cla.lecturer_id = ?", *lecturerID)
	}
	var items []LecturerCourseAssignment
	err := query.Order("courses.code ASC, users.last_name ASC").Scan(&items).Error
	return items, err
}

func (r *TeachingRepository) LecturerCourseMetrics(ctx context.Context, courseID, lecturerID uint) (LecturerCourseMetrics, error) {
	var metrics LecturerCourseMetrics
	db := r.db.WithContext(ctx)

	count := func(model any, where string, args ...any) (int64, error) {
		var value int64
		err := db.Model(model).Where(where, args...).Count(&value).Error
		return value, err
	}
	maxTime := func(model any, where string, args ...any) (*time.Time, error) {
		var value *time.Time
		err := db.Model(model).Select("MAX(updated_at)").Where(where, args...).Scan(&value).Error
		return value, err
	}

	var err error
	if metrics.AssignmentsPublished, err = count(&models.Assignment{}, "course_id = ? AND created_by = ? AND status <> ?", courseID, lecturerID, models.AssignmentStatusDraft); err != nil {
		return metrics, err
	}
	if metrics.AssignmentSubmissionCount, err = r.countAssignmentSubmissionsForCourse(ctx, courseID); err != nil {
		return metrics, err
	}
	if metrics.AssignmentsMarked, err = r.countAssignmentGradesForCourse(ctx, courseID); err != nil {
		return metrics, err
	}
	if metrics.ForumDiscussionCount, err = count(&models.CourseForumPost{}, "course_id = ?", courseID); err != nil {
		return metrics, err
	}
	if metrics.ExamCount, err = count(&models.Exam{}, "course_id = ? AND lecturer_id = ?", courseID, lecturerID); err != nil {
		return metrics, err
	}
	if err = db.Model(&models.ExamAttempt{}).
		Joins("JOIN exams ON exams.id = exam_attempts.exam_id").
		Where("exams.course_id = ? AND exams.lecturer_id = ? AND exam_attempts.status = ?", courseID, lecturerID, models.ExamAttemptMarked).
		Count(&metrics.ExamMarkedCount).Error; err != nil {
		return metrics, err
	}
	if err = db.Model(&models.ExamAttempt{}).
		Joins("JOIN exams ON exams.id = exam_attempts.exam_id").
		Select("AVG(exam_attempts.score)").
		Where("exams.course_id = ? AND exams.lecturer_id = ? AND exam_attempts.status = ?", courseID, lecturerID, models.ExamAttemptMarked).
		Scan(&metrics.AverageExamScore).Error; err != nil {
		return metrics, err
	}
	if metrics.LiveSessionsScheduled, err = count(&models.LiveSession{}, "course_id = ? AND created_by = ?", courseID, lecturerID); err != nil {
		return metrics, err
	}
	if metrics.LiveSessionsConducted, err = count(&models.LiveSession{}, "course_id = ? AND created_by = ? AND status = ?", courseID, lecturerID, models.LiveSessionStatusCompleted); err != nil {
		return metrics, err
	}
	if metrics.VideoLecturesUploaded, err = count(&models.VideoLecture{}, "course_id = ? AND uploaded_by = ?", courseID, lecturerID); err != nil {
		return metrics, err
	}
	if metrics.MaterialsUploaded, err = count(&models.CourseMaterial{}, "course_id = ? AND uploaded_by = ?", courseID, lecturerID); err != nil {
		return metrics, err
	}

	if metrics.LastAssignmentAt, err = maxTime(&models.Assignment{}, "course_id = ? AND created_by = ?", courseID, lecturerID); err != nil {
		return metrics, err
	}
	if metrics.LastMarkedAt, err = r.latestAssignmentGradeActivity(ctx, courseID, lecturerID); err != nil {
		return metrics, err
	}
	if metrics.LastForumAt, err = maxTime(&models.CourseForumPost{}, "course_id = ?", courseID); err != nil {
		return metrics, err
	}
	if metrics.LastExamAt, err = maxTime(&models.Exam{}, "course_id = ? AND lecturer_id = ?", courseID, lecturerID); err != nil {
		return metrics, err
	}
	if metrics.LastLiveSessionAt, err = maxTime(&models.LiveSession{}, "course_id = ? AND created_by = ?", courseID, lecturerID); err != nil {
		return metrics, err
	}
	if metrics.LastVideoLectureAt, err = maxTime(&models.VideoLecture{}, "course_id = ? AND uploaded_by = ?", courseID, lecturerID); err != nil {
		return metrics, err
	}
	if metrics.LastMaterialAt, err = maxTime(&models.CourseMaterial{}, "course_id = ? AND uploaded_by = ?", courseID, lecturerID); err != nil {
		return metrics, err
	}

	return metrics, nil
}

func (r *TeachingRepository) countAssignmentSubmissionsForCourse(ctx context.Context, courseID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("assignment_submissions").
		Joins("JOIN assignments ON assignments.id = assignment_submissions.assignment_id").
		Where("assignments.course_id = ?", courseID).
		Count(&count).Error
	return count, err
}

func (r *TeachingRepository) countAssignmentGradesForCourse(ctx context.Context, courseID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("assignment_grades").
		Joins("JOIN assignments ON assignments.id = assignment_grades.assignment_id").
		Where("assignments.course_id = ?", courseID).
		Count(&count).Error
	return count, err
}

func (r *TeachingRepository) latestAssignmentGradeActivity(ctx context.Context, courseID, lecturerID uint) (*time.Time, error) {
	var value *time.Time
	err := r.db.WithContext(ctx).
		Table("assignment_grades").
		Select("MAX(assignment_grades.updated_at)").
		Joins("JOIN assignments ON assignments.id = assignment_grades.assignment_id").
		Where("assignments.course_id = ? AND assignment_grades.marker_id = ?", courseID, lecturerID).
		Scan(&value).Error
	return value, err
}

func (r *TeachingRepository) countByCourse(ctx context.Context, model any, courseID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(model).Where("course_id = ?", courseID).Count(&count).Error
	return count, err
}

func (r *TeachingRepository) EnsureCourseBelongsToSession(ctx context.Context, sessionID, recordingID uint) error {
	var count int64
	err := r.db.WithContext(ctx).
		Table("live_session_recordings").
		Where("id = ? AND live_session_id = ?", recordingID, sessionID).
		Count(&count).Error
	if err != nil {
		return err
	}

	if count == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *TeachingRepository) DebugName() string {
	return fmt.Sprintf("teaching_repository:%p", r)
}
