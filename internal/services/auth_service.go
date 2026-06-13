package services

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"kslasbackend/internal/database/models"
	"kslasbackend/internal/dto"
	"kslasbackend/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInactiveAccount    = errors.New("account is not active")
)

type AuthService struct {
	repo            *repository.AuthRepository
	passwordService *PasswordService
	jwtService      *JWTService
}

func NewAuthService(repo *repository.AuthRepository, passwordService *PasswordService, jwtService *JWTService) *AuthService {
	return &AuthService{
		repo:            repo,
		passwordService: passwordService,
		jwtService:      jwtService,
	}
}

func (s *AuthService) Login(ctx context.Context, identity, password string) (*dto.LoginResponse, error) {
	identity = strings.TrimSpace(identity)
	if identity == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.repo.FindUserByIdentity(ctx, identity)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}

		return nil, err
	}

	if user.Status != models.UserStatusActive {
		return nil, ErrInactiveAccount
	}

	if err := s.passwordService.Verify(password, user.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	now := time.Now().UTC()
	if err := s.repo.UpdateLastLogin(ctx, user.ID, now); err != nil {
		return nil, err
	}
	user.LastLoginAt = &now

	token, err := s.jwtService.GenerateToken(user)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:      token,
		TokenType:        "Bearer",
		ExpiresInSeconds: s.jwtService.ExpiresInSeconds(),
		User:             buildUserPayload(user),
	}, nil
}

func (s *AuthService) CurrentUser(ctx context.Context, userID uint) (*dto.UserPayload, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	payload := buildUserPayload(user)
	return &payload, nil
}

func buildUserPayload(user *models.User) dto.UserPayload {
	now := time.Now().UTC()
	roles := make([]dto.UserRolePayload, 0, len(user.UserRoles))

	for _, assignment := range user.UserRoles {
		if !assignment.IsActiveAt(now) {
			continue
		}

		roles = append(roles, dto.UserRolePayload{
			Code:      assignment.Role.Code,
			Name:      assignment.Role.Name,
			ScopeType: string(assignment.ScopeType),
			ScopeID:   assignment.ScopeID,
			IsPrimary: assignment.IsPrimary,
		})
	}

	sort.SliceStable(roles, func(i, j int) bool {
		if roles[i].IsPrimary != roles[j].IsPrimary {
			return roles[i].IsPrimary
		}

		if roles[i].Code != roles[j].Code {
			return roles[i].Code < roles[j].Code
		}

		if roles[i].ScopeType != roles[j].ScopeType {
			return roles[i].ScopeType < roles[j].ScopeType
		}

		switch {
		case roles[i].ScopeID == nil && roles[j].ScopeID == nil:
			return false
		case roles[i].ScopeID == nil:
			return true
		case roles[j].ScopeID == nil:
			return false
		default:
			return *roles[i].ScopeID < *roles[j].ScopeID
		}
	})

	return dto.UserPayload{
		ID:          user.ID,
		UUID:        user.UUID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		MiddleName:  user.MiddleName,
		Email:       user.Email,
		Phone:       user.Phone,
		UserType:    string(user.UserType),
		Status:      string(user.Status),
		LastLoginAt: user.LastLoginAt,
		Roles:       roles,
	}
}
