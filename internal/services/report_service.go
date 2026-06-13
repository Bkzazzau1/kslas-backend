package services

import (
	"context"
	"fmt"
	"strings"

	"kslasbackend/internal/dto"
	"kslasbackend/internal/rbac"
	"kslasbackend/internal/repository"
)

type ReportService struct {
	repo              *repository.TeachingRepository
	permissionService *PermissionService
}

func NewReportService(repo *repository.TeachingRepository, permissionService *PermissionService) *ReportService {
	return &ReportService{repo: repo, permissionService: permissionService}
}

func (s *ReportService) LecturerReports(ctx context.Context, userID uint, courseID, lecturerID *uint) ([]dto.LecturerCourseReportResponse, error) {
	if err := s.ensurePermission(ctx, userID, "report.department.view", scopePtr(rbac.SchoolScope())); err != nil {
		return nil, err
	}
	assignments, err := s.repo.ListLecturerCourseAssignments(ctx, courseID, lecturerID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.LecturerCourseReportResponse, 0, len(assignments))
	for _, item := range assignments {
		metrics, err := s.repo.LecturerCourseMetrics(ctx, item.CourseID, item.LecturerID)
		if err != nil {
			return nil, err
		}
		markingPercent := percent(metrics.AssignmentsMarked, metrics.AssignmentSubmissionCount)
		health := "needs_attention"
		if markingPercent >= 80 && metrics.LiveSessionsScheduled > 0 && (metrics.AssignmentsPublished > 0 || metrics.ExamCount > 0) {
			health = "ok"
		}
		out = append(out, dto.LecturerCourseReportResponse{
			CourseID:                  item.CourseID,
			CourseCode:                item.CourseCode,
			CourseTitle:               item.CourseTitle,
			LecturerID:                item.LecturerID,
			LecturerName:              strings.TrimSpace(strings.TrimSpace(item.FirstName) + " " + strings.TrimSpace(item.LastName)),
			AssignmentsPublished:      metrics.AssignmentsPublished,
			AssignmentSubmissionCount: metrics.AssignmentSubmissionCount,
			AssignmentsMarked:         metrics.AssignmentsMarked,
			AssignmentMarkingPercent:  markingPercent,
			ForumDiscussionCount:      metrics.ForumDiscussionCount,
			ExamCount:                 metrics.ExamCount,
			ExamMarkedCount:           metrics.ExamMarkedCount,
			AverageExamScore:          metrics.AverageExamScore,
			LiveSessionsScheduled:     metrics.LiveSessionsScheduled,
			LiveSessionsConducted:     metrics.LiveSessionsConducted,
			VideoLecturesUploaded:     metrics.VideoLecturesUploaded,
			MaterialsUploaded:         metrics.MaterialsUploaded,
			HealthStatus:              health,
			Activities:                lecturerActivities(metrics, markingPercent),
		})
	}
	return out, nil
}

func (s *ReportService) ensurePermission(ctx context.Context, userID uint, permissionCode string, target *rbac.Scope) error {
	allowed, err := s.permissionService.UserHasPermission(ctx, userID, permissionCode, target)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrPermissionDenied
	}
	return nil
}

func percent(part, total int64) float64 {
	if total <= 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}

func lecturerActivities(metrics repository.LecturerCourseMetrics, markingPercent float64) []dto.LecturerActivityItem {
	items := []dto.LecturerActivityItem{}
	if metrics.AssignmentsPublished > 0 {
		items = append(items, dto.LecturerActivityItem{
			Type:       "assignments_published",
			Summary:    fmt.Sprintf("%d assignments published", metrics.AssignmentsPublished),
			OccurredAt: metrics.LastAssignmentAt,
		})
	}
	if metrics.AssignmentSubmissionCount > 0 {
		items = append(items, dto.LecturerActivityItem{
			Type:       "assignment_marking",
			Summary:    fmt.Sprintf("students assignments %.0f%% marked", markingPercent),
			OccurredAt: metrics.LastMarkedAt,
		})
	}
	if metrics.ExamCount > 0 {
		items = append(items, dto.LecturerActivityItem{
			Type:       "exam_questions_submitted",
			Summary:    fmt.Sprintf("%d exam question sets submitted to exam officer", metrics.ExamCount),
			OccurredAt: metrics.LastExamAt,
		})
	}
	if metrics.LiveSessionsScheduled > 0 {
		items = append(items, dto.LecturerActivityItem{
			Type:       "live_sessions",
			Summary:    fmt.Sprintf("%d live sessions scheduled, %d conducted", metrics.LiveSessionsScheduled, metrics.LiveSessionsConducted),
			OccurredAt: metrics.LastLiveSessionAt,
		})
	}
	if metrics.ForumDiscussionCount > 0 {
		items = append(items, dto.LecturerActivityItem{
			Type:       "forum_discussion",
			Summary:    fmt.Sprintf("%d course forum discussions/replies", metrics.ForumDiscussionCount),
			OccurredAt: metrics.LastForumAt,
		})
	}
	if metrics.VideoLecturesUploaded > 0 {
		items = append(items, dto.LecturerActivityItem{
			Type:       "video_lectures_uploaded",
			Summary:    fmt.Sprintf("%d video lectures uploaded", metrics.VideoLecturesUploaded),
			OccurredAt: metrics.LastVideoLectureAt,
		})
	}
	if metrics.MaterialsUploaded > 0 {
		items = append(items, dto.LecturerActivityItem{
			Type:       "materials_uploaded",
			Summary:    fmt.Sprintf("%d course materials uploaded", metrics.MaterialsUploaded),
			OccurredAt: metrics.LastMaterialAt,
		})
	}
	return items
}
