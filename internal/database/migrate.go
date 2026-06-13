package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"kslasbackend/internal/config"
	"kslasbackend/internal/database/models"
	"kslasbackend/internal/database/seeds"
	"kslasbackend/internal/services"
)

func Bootstrap(ctx context.Context, db *gorm.DB, cfg config.Config) error {
	if cfg.AutoMigrate {
		if err := AutoMigrate(db.WithContext(ctx)); err != nil {
			return fmt.Errorf("auto migrate: %w", err)
		}
	}

	if cfg.SeedRBAC {
		if err := seeds.SeedRBAC(ctx, db); err != nil {
			return fmt.Errorf("seed rbac: %w", err)
		}
	}

	if err := SeedBootstrapAdmin(ctx, db, cfg, services.NewPasswordService()); err != nil {
		return fmt.Errorf("seed bootstrap admin: %w", err)
	}

	return nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.Faculty{},
		&models.Department{},
		&models.Programme{},
		&models.Course{},
		&models.CourseLecturerAssignment{},
		&models.StudentAcademicProfile{},
		&models.CourseRegistration{},
		&models.CourseMaterial{},
		&models.VideoLecture{},
		&models.VideoLectureWatch{},
		&models.LiveSession{},
		&models.LiveSessionAttendance{},
		&models.LiveSessionRecording{},
		&models.CourseForumPost{},
		&models.CourseDirectMessage{},
		&models.Assignment{},
		&models.AssignmentGroup{},
		&models.AssignmentGroupMember{},
		&models.AssignmentSubmission{},
		&models.AssignmentSubmissionFile{},
		&models.AssignmentPeerReview{},
		&models.AssignmentGrade{},
		&models.Quiz{},
		&models.ExamVenue{},
		&models.Exam{},
		&models.ExamStudentAllocation{},
		&models.ExamInvigilator{},
		&models.ExamAttempt{},
		&models.ExamScriptAnnotation{},
		&models.ProctoringAlert{},
		&models.Result{},
		&models.RolePermission{},
		&models.UserRole{},
	)
}
