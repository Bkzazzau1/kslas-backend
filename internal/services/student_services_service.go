package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"kslasbackend/internal/database/models"
	"kslasbackend/internal/dto"
	"kslasbackend/internal/repository"
)

type StudentServicesService struct {
	repository *repository.StudentServicesRepository
}

func NewStudentServicesService(repository *repository.StudentServicesRepository) *StudentServicesService {
	return &StudentServicesService{repository: repository}
}

func (s *StudentServicesService) GetGraduationMap(ctx context.Context, studentID uint) (*dto.GraduationMapResponse, error) {
	graduationMap, err := s.repository.FindGraduationMapByStudent(ctx, studentID)
	if err != nil {
		return nil, err
	}

	items, err := s.repository.ListGraduationMapItems(ctx, graduationMap.ID)
	if err != nil {
		return nil, err
	}

	response := &dto.GraduationMapResponse{
		StudentID:          graduationMap.StudentID,
		ProgrammeID:        graduationMap.ProgrammeID,
		CreditsRequired:    graduationMap.CreditsRequired,
		CreditsEarned:      graduationMap.CreditsEarned,
		CreditsInProgress:  graduationMap.CreditsInProgress,
		CreditsRemaining:   graduationMap.CreditsRemaining,
		CoursesCompleted:   graduationMap.CoursesCompleted,
		CoursesRemaining:   graduationMap.CoursesRemaining,
		Carryovers:         graduationMap.Carryovers,
		ProgressPercent:    graduationMap.ProgressPercent,
		ExpectedGraduation: graduationMap.ExpectedGraduation,
		GeneratedAt:        graduationMap.GeneratedAt,
		Items:              make([]dto.GraduationMapItemResponse, 0, len(items)),
	}

	for _, item := range items {
		response.Items = append(response.Items, dto.GraduationMapItemResponse{
			CourseID:        item.CourseID,
			Code:            item.Code,
			Title:           item.Title,
			RequirementType: item.RequirementType,
			Credits:         item.Credits,
			Status:          item.Status,
		})
	}

	return response, nil
}

func (s *StudentServicesService) ListSupportTickets(ctx context.Context, studentID uint) ([]dto.SupportTicketResponse, error) {
	tickets, err := s.repository.ListSupportTicketsByStudent(ctx, studentID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.SupportTicketResponse, 0, len(tickets))
	for _, ticket := range tickets {
		responses = append(responses, mapSupportTicket(ticket))
	}
	return responses, nil
}

func (s *StudentServicesService) CreateSupportTicket(ctx context.Context, studentID uint, request dto.CreateSupportTicketRequest) (*dto.SupportTicketResponse, error) {
	category := strings.TrimSpace(request.Category)
	subject := strings.TrimSpace(request.Subject)
	description := strings.TrimSpace(request.Description)
	priority := strings.ToLower(strings.TrimSpace(request.Priority))

	if category == "" || subject == "" || description == "" {
		return nil, errors.New("category, subject and description are required")
	}

	if priority == "" {
		priority = "normal"
	}

	ticket := &models.SupportTicket{
		TicketNo:    buildTicketNumber(studentID),
		StudentID:   studentID,
		Category:    category,
		Subject:     subject,
		Description: description,
		Priority:    priority,
		Status:      models.SupportTicketOpen,
	}

	if err := s.repository.CreateSupportTicket(ctx, ticket); err != nil {
		return nil, err
	}

	response := mapSupportTicket(*ticket)
	return &response, nil
}

func (s *StudentServicesService) AddSupportReply(ctx context.Context, studentID uint, ticketID uint, request dto.AddSupportReplyRequest) error {
	message := strings.TrimSpace(request.Message)
	if message == "" {
		return errors.New("message is required")
	}

	if _, err := s.repository.FindSupportTicketForStudent(ctx, ticketID, studentID); err != nil {
		return err
	}

	return s.repository.AddSupportTicketMessage(ctx, &models.SupportTicketMessage{
		TicketID:   ticketID,
		SenderID:   studentID,
		SenderType: "student",
		Message:    message,
	})
}

func (s *StudentServicesService) GetInternshipProfile(ctx context.Context, studentID uint) (*dto.InternshipProfileResponse, error) {
	profile, err := s.repository.FindInternshipProfileByStudent(ctx, studentID)
	if err != nil {
		return nil, err
	}

	response := mapInternshipProfile(*profile)
	return &response, nil
}

func (s *StudentServicesService) UpsertInternshipProfile(ctx context.Context, studentID uint, request dto.UpsertInternshipProfileRequest) (*dto.InternshipProfileResponse, error) {
	if request.ProgrammeID == 0 {
		return nil, errors.New("programme_id is required")
	}

	profile, err := s.repository.FindInternshipProfileByStudent(ctx, studentID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		profile = &models.InternshipProfile{StudentID: studentID, Status: models.InternshipProfileDraft}
	}

	profile.ProgrammeID = request.ProgrammeID
	profile.PreferredIndustry = request.PreferredIndustry
	profile.OrganizationName = request.OrganizationName
	profile.OrganizationEmail = request.OrganizationEmail
	profile.OrganizationPhone = request.OrganizationPhone
	profile.OrganizationAddr = request.OrganizationAddr
	profile.StartDate = request.StartDate
	profile.EndDate = request.EndDate

	if err := s.repository.UpsertInternshipProfile(ctx, profile); err != nil {
		return nil, err
	}

	response := mapInternshipProfile(*profile)
	return &response, nil
}

func (s *StudentServicesService) ListTranscriptRequests(ctx context.Context, studentID uint) ([]dto.TranscriptRequestResponse, error) {
	requests, err := s.repository.ListTranscriptRequestsByStudent(ctx, studentID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.TranscriptRequestResponse, 0, len(requests))
	for _, request := range requests {
		responses = append(responses, mapTranscriptRequest(request))
	}
	return responses, nil
}

func (s *StudentServicesService) CreateTranscriptRequest(ctx context.Context, studentID uint, request dto.CreateTranscriptRequestRequest) (*dto.TranscriptRequestResponse, error) {
	deliveryMethod := strings.TrimSpace(request.DeliveryMethod)
	recipientName := strings.TrimSpace(request.RecipientName)
	recipientContact := strings.TrimSpace(request.RecipientContact)
	purpose := strings.TrimSpace(request.Purpose)

	if deliveryMethod == "" || recipientName == "" || recipientContact == "" || purpose == "" {
		return nil, errors.New("delivery_method, recipient_name, recipient_contact and purpose are required")
	}

	transcriptRequest := &models.TranscriptRequest{
		RequestNo:        buildTranscriptRequestNumber(studentID),
		StudentID:        studentID,
		DeliveryMethod:   deliveryMethod,
		RecipientName:    recipientName,
		RecipientContact: recipientContact,
		Purpose:          purpose,
		Status:           "pending",
	}

	if err := s.repository.CreateTranscriptRequest(ctx, transcriptRequest); err != nil {
		return nil, err
	}

	response := mapTranscriptRequest(*transcriptRequest)
	return &response, nil
}

func mapSupportTicket(ticket models.SupportTicket) dto.SupportTicketResponse {
	return dto.SupportTicketResponse{
		ID:             ticket.ID,
		TicketNo:       ticket.TicketNo,
		Category:       ticket.Category,
		Subject:        ticket.Subject,
		Description:    ticket.Description,
		Priority:       ticket.Priority,
		Status:         string(ticket.Status),
		CurrentOwnerID: ticket.CurrentOwnerID,
		CreatedAt:      ticket.CreatedAt,
		UpdatedAt:      ticket.UpdatedAt,
		ResolvedAt:     ticket.ResolvedAt,
		ClosedAt:       ticket.ClosedAt,
	}
}

func mapInternshipProfile(profile models.InternshipProfile) dto.InternshipProfileResponse {
	return dto.InternshipProfileResponse{
		ID:                profile.ID,
		StudentID:         profile.StudentID,
		ProgrammeID:       profile.ProgrammeID,
		PreferredIndustry: profile.PreferredIndustry,
		OrganizationName:  profile.OrganizationName,
		OrganizationEmail: profile.OrganizationEmail,
		OrganizationPhone: profile.OrganizationPhone,
		OrganizationAddr:  profile.OrganizationAddr,
		StartDate:         profile.StartDate,
		EndDate:           profile.EndDate,
		Status:            string(profile.Status),
		CreatedAt:         profile.CreatedAt,
		UpdatedAt:         profile.UpdatedAt,
	}
}

func mapTranscriptRequest(request models.TranscriptRequest) dto.TranscriptRequestResponse {
	return dto.TranscriptRequestResponse{
		ID:               request.ID,
		RequestNo:        request.RequestNo,
		DeliveryMethod:   request.DeliveryMethod,
		RecipientName:    request.RecipientName,
		RecipientContact: request.RecipientContact,
		Purpose:          request.Purpose,
		Status:           request.Status,
		RequestedAt:      request.RequestedAt,
		CompletedAt:      request.CompletedAt,
	}
}

func buildTicketNumber(studentID uint) string {
	return fmt.Sprintf("KSLAS-SUP-%d-%d", studentID, time.Now().UTC().UnixNano())
}

func buildTranscriptRequestNumber(studentID uint) string {
	return fmt.Sprintf("TRX-%d-%d", studentID, time.Now().UTC().UnixNano())
}
