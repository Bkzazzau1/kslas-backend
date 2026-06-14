package services

import (
	"context"
	"encoding/json"
	"fmt"

	"kslasbackend/internal/database/models"
	"kslasbackend/internal/repository"
)

type AuditService struct {
	repository *repository.AuditRepository
}

type AuditRecordInput struct {
	ActorUserID uint
	ActorRole   string
	Action      string
	EntityType  string
	EntityID    string
	Before      any
	After       any
	IPAddress   string
	UserAgent   string
}

func NewAuditService(repository *repository.AuditRepository) *AuditService {
	return &AuditService{repository: repository}
}

func (s *AuditService) Record(ctx context.Context, input AuditRecordInput) error {
	beforeJSON, err := marshalAuditPayload(input.Before)
	if err != nil {
		return fmt.Errorf("marshal audit before payload: %w", err)
	}

	afterJSON, err := marshalAuditPayload(input.After)
	if err != nil {
		return fmt.Errorf("marshal audit after payload: %w", err)
	}

	log := &models.AuditLog{
		ActorUserID: input.ActorUserID,
		ActorRole:   input.ActorRole,
		Action:      input.Action,
		EntityType:  input.EntityType,
		EntityID:    input.EntityID,
		BeforeJSON:  beforeJSON,
		AfterJSON:   afterJSON,
		IPAddress:   input.IPAddress,
		UserAgent:   input.UserAgent,
	}

	return s.repository.Create(ctx, log)
}

func marshalAuditPayload(value any) (string, error) {
	if value == nil {
		return "{}", nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
