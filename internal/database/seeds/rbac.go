package seeds

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"kslasbackend/internal/database/models"
)

func SeedRBAC(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := seedRoles(tx); err != nil {
			return err
		}

		if err := seedPermissions(tx); err != nil {
			return err
		}

		if err := seedRolePermissions(tx); err != nil {
			return err
		}

		return nil
	})
}

func seedRoles(tx *gorm.DB) error {
	for _, role := range roleCatalog() {
		record := models.Role{Code: role.Code}
		if err := tx.Where("code = ?", role.Code).
			Assign(map[string]any{
				"name":        role.Name,
				"description": role.Description,
				"is_system":   role.IsSystem,
			}).
			FirstOrCreate(&record).Error; err != nil {
			return fmt.Errorf("seed role %s: %w", role.Code, err)
		}
	}

	return nil
}

func seedPermissions(tx *gorm.DB) error {
	for _, permission := range permissionCatalog() {
		record := models.Permission{Code: permission.Code}
		if err := tx.Where("code = ?", permission.Code).
			Assign(map[string]any{
				"name":        permission.Name,
				"module":      permission.Module,
				"description": permission.Description,
			}).
			FirstOrCreate(&record).Error; err != nil {
			return fmt.Errorf("seed permission %s: %w", permission.Code, err)
		}
	}

	return nil
}

func seedRolePermissions(tx *gorm.DB) error {
	var roles []models.Role
	if err := tx.Find(&roles).Error; err != nil {
		return fmt.Errorf("load roles: %w", err)
	}

	roleIDByCode := make(map[string]uint, len(roles))
	for _, role := range roles {
		roleIDByCode[role.Code] = role.ID
	}

	var permissions []models.Permission
	if err := tx.Find(&permissions).Error; err != nil {
		return fmt.Errorf("load permissions: %w", err)
	}

	permissionIDByCode := make(map[string]uint, len(permissions))
	allPermissionCodes := make([]string, 0, len(permissions))
	for _, permission := range permissions {
		permissionIDByCode[permission.Code] = permission.ID
		allPermissionCodes = append(allPermissionCodes, permission.Code)
	}

	assignments := rolePermissionCatalog(allPermissionCodes)

	for roleCode, permissionCodes := range assignments {
		roleID, ok := roleIDByCode[roleCode]
		if !ok {
			return fmt.Errorf("role %s not found during role-permission seed", roleCode)
		}

		for _, permissionCode := range permissionCodes {
			permissionID, ok := permissionIDByCode[permissionCode]
			if !ok {
				return fmt.Errorf("permission %s not found during role-permission seed", permissionCode)
			}

			record := models.RolePermission{
				RoleID:       roleID,
				PermissionID: permissionID,
			}

			if err := tx.Where("role_id = ? AND permission_id = ?", roleID, permissionID).
				FirstOrCreate(&record).Error; err != nil {
				return fmt.Errorf("seed role permission %s -> %s: %w", roleCode, permissionCode, err)
			}
		}
	}

	return nil
}

func roleCatalog() []models.Role {
	return []models.Role{
		{Name: "System Admin", Code: "system_admin", Description: "Manages the entire smart learning platform.", IsSystem: true},
		{Name: "Academic Admin", Code: "academic_admin", Description: "Coordinates academic setup across the school.", IsSystem: true},
		{Name: "Dean", Code: "dean", Description: "Oversees a faculty and its academic operations.", IsSystem: true},
		{Name: "Head of Department", Code: "hod", Description: "Leads a department and its course operations.", IsSystem: true},
		{Name: "Programme Coordinator", Code: "programme_coordinator", Description: "Coordinates a programme and related course delivery.", IsSystem: true},
		{Name: "Exam Officer", Code: "exam_officer", Description: "Manages exam scheduling and exam operations.", IsSystem: true},
		{Name: "Lecturer", Code: "lecturer", Description: "Teaches and manages assigned courses.", IsSystem: true},
		{Name: "Teaching Assistant", Code: "teaching_assistant", Description: "Supports course delivery and marking.", IsSystem: true},
		{Name: "Moderator", Code: "moderator", Description: "Reviews and approves assessments and results.", IsSystem: true},
		{Name: "Proctor", Code: "proctor", Description: "Monitors and proctors exams.", IsSystem: true},
		{Name: "Student", Code: "student", Description: "Learns, submits work, and views results.", IsSystem: true},
		{Name: "Content Manager", Code: "content_manager", Description: "Maintains course materials and digital content.", IsSystem: true},
		{Name: "Student Affairs", Code: "student_affairs", Description: "Supports student-facing administrative workflows.", IsSystem: true},
		{Name: "Registry Officer", Code: "registry_officer", Description: "Registers students and manages course registration records.", IsSystem: true},
		{Name: "Marker", Code: "marker", Description: "Marks assessments and enters results.", IsSystem: true},
		{Name: "Class Rep", Code: "class_rep", Description: "Represents a class and accesses limited course information.", IsSystem: true},
	}
}

func permissionCatalog() []models.Permission {
	return []models.Permission{
		{Name: "Create User", Code: "user.create", Module: "user", Description: "Create a new platform user."},
		{Name: "View User", Code: "user.view", Module: "user", Description: "View user records."},
		{Name: "Update User", Code: "user.update", Module: "user", Description: "Update existing users."},
		{Name: "Delete User", Code: "user.delete", Module: "user", Description: "Delete or deactivate users."},
		{Name: "Assign Role", Code: "user.assign_role", Module: "user", Description: "Assign RBAC roles to users."},

		{Name: "Create Faculty", Code: "faculty.create", Module: "faculty", Description: "Create faculty records."},
		{Name: "View Faculty", Code: "faculty.view", Module: "faculty", Description: "View faculty records."},
		{Name: "Update Faculty", Code: "faculty.update", Module: "faculty", Description: "Update faculty records."},
		{Name: "Delete Faculty", Code: "faculty.delete", Module: "faculty", Description: "Delete faculty records."},

		{Name: "Create Department", Code: "department.create", Module: "department", Description: "Create department records."},
		{Name: "View Department", Code: "department.view", Module: "department", Description: "View department records."},
		{Name: "Update Department", Code: "department.update", Module: "department", Description: "Update department records."},
		{Name: "Delete Department", Code: "department.delete", Module: "department", Description: "Delete department records."},

		{Name: "Create Programme", Code: "programme.create", Module: "programme", Description: "Create programme records."},
		{Name: "View Programme", Code: "programme.view", Module: "programme", Description: "View programme records."},
		{Name: "Update Programme", Code: "programme.update", Module: "programme", Description: "Update programme records."},
		{Name: "Delete Programme", Code: "programme.delete", Module: "programme", Description: "Delete programme records."},

		{Name: "Create Course", Code: "course.create", Module: "course", Description: "Create course records."},
		{Name: "View Course", Code: "course.view", Module: "course", Description: "View course records."},
		{Name: "Update Course", Code: "course.update", Module: "course", Description: "Update course records."},
		{Name: "Delete Course", Code: "course.delete", Module: "course", Description: "Delete course records."},
		{Name: "Assign Lecturer", Code: "course.assign_lecturer", Module: "course", Description: "Assign lecturers to courses."},
		{Name: "Register Course", Code: "course.register", Module: "course", Description: "Register students for eligible courses."},
		{Name: "Approve Course Registration", Code: "course.registration.approve", Module: "course", Description: "Approve student course registrations."},
		{Name: "Manage Course Content", Code: "course.manage_content", Module: "course", Description: "Manage learning content for a course."},

		{Name: "Upload Material", Code: "material.upload", Module: "material", Description: "Upload course materials."},
		{Name: "View Material", Code: "material.view", Module: "material", Description: "View course materials."},
		{Name: "Create Live Class", Code: "liveclass.create", Module: "liveclass", Description: "Create a live class session."},
		{Name: "Join Live Class", Code: "liveclass.join", Module: "liveclass", Description: "Join a live class session."},
		{Name: "Manage Live Class", Code: "liveclass.manage", Module: "liveclass", Description: "Manage a live class session."},
		{Name: "Take Attendance", Code: "attendance.take", Module: "attendance", Description: "Take attendance for a class."},
		{Name: "View Attendance", Code: "attendance.view", Module: "attendance", Description: "View attendance records."},

		{Name: "Create Assignment", Code: "assignment.create", Module: "assignment", Description: "Create assignments."},
		{Name: "Submit Assignment", Code: "assignment.submit", Module: "assignment", Description: "Submit assignments."},
		{Name: "Mark Assignment", Code: "assignment.mark", Module: "assignment", Description: "Mark assignments."},
		{Name: "Create Quiz", Code: "quiz.create", Module: "quiz", Description: "Create quizzes."},
		{Name: "Submit Quiz", Code: "quiz.submit", Module: "quiz", Description: "Submit quizzes."},
		{Name: "Approve Assessment", Code: "assessment.approve", Module: "assessment", Description: "Approve assessments before release."},

		{Name: "Create Exam", Code: "exam.create", Module: "exam", Description: "Create exam records."},
		{Name: "Schedule Exam", Code: "exam.schedule", Module: "exam", Description: "Schedule exams."},
		{Name: "Start Exam", Code: "exam.start", Module: "exam", Description: "Start an exam session."},
		{Name: "Monitor Exam", Code: "exam.monitor", Module: "exam", Description: "Monitor exam activity."},
		{Name: "Proctor Exam", Code: "exam.proctor", Module: "exam", Description: "Proctor exams."},
		{Name: "Submit Exam Review", Code: "exam.submit_review", Module: "exam", Description: "Submit an exam moderation review."},

		{Name: "View Result", Code: "result.view", Module: "result", Description: "View results."},
		{Name: "Mark Result", Code: "result.mark", Module: "result", Description: "Enter or mark results."},
		{Name: "Approve Result", Code: "result.approve", Module: "result", Description: "Approve results."},
		{Name: "Publish Result", Code: "result.publish", Module: "result", Description: "Publish results."},
		{Name: "Generate Transcript", Code: "transcript.generate", Module: "transcript", Description: "Generate student transcripts."},

		{Name: "View School Reports", Code: "report.school.view", Module: "report", Description: "View school-wide reports."},
		{Name: "View Faculty Reports", Code: "report.faculty.view", Module: "report", Description: "View faculty reports."},
		{Name: "View Department Reports", Code: "report.department.view", Module: "report", Description: "View department reports."},
		{Name: "View Course Reports", Code: "report.course.view", Module: "report", Description: "View course reports."},
	}
}

func rolePermissionCatalog(allPermissionCodes []string) map[string][]string {
	return map[string][]string{
		"system_admin": allPermissionCodes,
		"academic_admin": {
			"user.create", "user.view", "user.update", "user.assign_role",
			"faculty.create", "faculty.view", "faculty.update", "faculty.delete",
			"department.create", "department.view", "department.update", "department.delete",
			"programme.create", "programme.view", "programme.update", "programme.delete",
			"course.create", "course.view", "course.update", "course.delete", "course.assign_lecturer",
			"exam.create", "exam.schedule", "result.approve", "result.publish",
			"report.school.view",
		},
		"dean": {
			"faculty.view", "faculty.update",
			"department.create", "department.view", "department.update", "department.delete",
			"programme.create", "programme.view", "programme.update", "programme.delete",
			"course.create", "course.view", "course.update", "course.delete", "course.assign_lecturer",
			"assessment.approve", "result.approve", "report.faculty.view",
		},
		"hod": {
			"user.create", "user.view", "user.assign_role",
			"department.view", "department.update",
			"programme.create", "programme.view", "programme.update", "programme.delete",
			"course.create", "course.view", "course.update", "course.delete", "course.assign_lecturer",
			"assessment.approve", "result.approve", "report.department.view",
		},
		"programme_coordinator": {
			"programme.view", "programme.update", "course.view", "course.update", "course.manage_content",
			"attendance.view", "result.view", "report.department.view",
		},
		"exam_officer": {
			"course.view", "exam.create", "exam.schedule", "exam.start",
			"exam.monitor", "exam.submit_review", "result.view", "result.mark",
			"result.publish", "report.school.view",
		},
		"lecturer": {
			"course.view", "course.manage_content", "material.upload", "material.view",
			"liveclass.create", "liveclass.manage", "attendance.take", "attendance.view",
			"assignment.create", "assignment.mark", "quiz.create",
			"exam.create", "result.mark", "report.course.view",
		},
		"teaching_assistant": {
			"course.view", "material.upload", "material.view", "liveclass.join",
			"attendance.view", "assignment.mark", "quiz.create", "report.course.view",
		},
		"moderator": {
			"course.view", "assessment.approve", "exam.submit_review",
			"result.approve", "report.department.view",
		},
		"proctor": {
			"course.view", "exam.monitor", "exam.proctor", "report.course.view",
		},
		"content_manager": {
			"course.view", "course.manage_content", "material.upload", "material.view",
		},
		"student_affairs": {
			"user.view", "result.view", "transcript.generate", "report.school.view",
		},
		"registry_officer": {
			"user.create", "user.view", "user.update", "user.assign_role",
			"department.view", "programme.view", "course.view",
			"course.register", "course.registration.approve", "result.view", "transcript.generate",
		},
		"student": {
			"course.view", "course.register", "material.view", "liveclass.join",
			"assignment.submit", "quiz.submit", "result.view",
		},
		"marker": {
			"course.view", "assignment.mark", "result.mark", "report.course.view",
		},
		"class_rep": {
			"course.view", "material.view", "liveclass.join", "attendance.view",
		},
	}
}
