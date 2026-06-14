package database

import "kslasbackend/internal/database/models"

func StudentServicesMigrationModels() []any {
	return []any{
		&models.AuditLog{},
		&models.SupportTicket{},
		&models.SupportTicketMessage{},
		&models.SupportTicketFile{},
		&models.InternshipProfile{},
		&models.InternshipLetterRequest{},
		&models.InternshipLogbookEntry{},
		&models.TranscriptRequest{},
		&models.GraduationMap{},
		&models.GraduationMapItem{},
	}
}
