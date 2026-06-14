package repository

import (
	"context"

	"gorm.io/gorm"

	"kslasbackend/internal/database/models"
)

type StudentServicesRepository struct {
	db *gorm.DB
}

func NewStudentServicesRepository(db *gorm.DB) *StudentServicesRepository {
	return &StudentServicesRepository{db: db}
}

func (r *StudentServicesRepository) FindGraduationMapByStudent(ctx context.Context, studentID uint) (*models.GraduationMap, error) {
	var graduationMap models.GraduationMap
	err := r.db.WithContext(ctx).
		Preload("Programme").
		Preload("Student").
		Where("student_id = ?", studentID).
		First(&graduationMap).Error
	if err != nil {
		return nil, err
	}
	return &graduationMap, nil
}

func (r *StudentServicesRepository) ListGraduationMapItems(ctx context.Context, graduationMapID uint) ([]models.GraduationMapItem, error) {
	var items []models.GraduationMapItem
	err := r.db.WithContext(ctx).
		Where("graduation_map_id = ?", graduationMapID).
		Order("status ASC, code ASC").
		Find(&items).Error
	return items, err
}

func (r *StudentServicesRepository) ListSupportTicketsByStudent(ctx context.Context, studentID uint) ([]models.SupportTicket, error) {
	var tickets []models.SupportTicket
	err := r.db.WithContext(ctx).
		Where("student_id = ?", studentID).
		Order("updated_at DESC").
		Find(&tickets).Error
	return tickets, err
}

func (r *StudentServicesRepository) CreateSupportTicket(ctx context.Context, ticket *models.SupportTicket) error {
	return r.db.WithContext(ctx).Create(ticket).Error
}

func (r *StudentServicesRepository) FindSupportTicketForStudent(ctx context.Context, ticketID uint, studentID uint) (*models.SupportTicket, error) {
	var ticket models.SupportTicket
	err := r.db.WithContext(ctx).
		Where("id = ? AND student_id = ?", ticketID, studentID).
		First(&ticket).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *StudentServicesRepository) AddSupportTicketMessage(ctx context.Context, message *models.SupportTicketMessage) error {
	return r.db.WithContext(ctx).Create(message).Error
}

func (r *StudentServicesRepository) FindInternshipProfileByStudent(ctx context.Context, studentID uint) (*models.InternshipProfile, error) {
	var profile models.InternshipProfile
	err := r.db.WithContext(ctx).
		Preload("Programme").
		Where("student_id = ?", studentID).
		First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *StudentServicesRepository) UpsertInternshipProfile(ctx context.Context, profile *models.InternshipProfile) error {
	return r.db.WithContext(ctx).Save(profile).Error
}

func (r *StudentServicesRepository) ListTranscriptRequestsByStudent(ctx context.Context, studentID uint) ([]models.TranscriptRequest, error) {
	var requests []models.TranscriptRequest
	err := r.db.WithContext(ctx).
		Where("student_id = ?", studentID).
		Order("requested_at DESC").
		Find(&requests).Error
	return requests, err
}

func (r *StudentServicesRepository) CreateTranscriptRequest(ctx context.Context, request *models.TranscriptRequest) error {
	return r.db.WithContext(ctx).Create(request).Error
}
