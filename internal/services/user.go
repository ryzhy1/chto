package services

import (
	"boton-back/internal/domain/dto"
	"boton-back/internal/repository"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"log/slog"
	"time"
)

type UserService struct {
	log            *slog.Logger
	userRepository UserRepository
	tokenTTL       time.Duration
}

type UserRepository interface {
	GetUser(ctx context.Context, userId uuid.UUID) (*dto.User, error)
	GetAllFriends(ctx context.Context, userId uuid.UUID) ([]*dto.Friend, error)
}

// NewUserService return a new instance of the Auth service
func NewUserService(log *slog.Logger, userRepository UserRepository) *UserService {
	return &UserService{
		log:            log,
		userRepository: userRepository,
	}
}

func (s *UserService) GetUser(ctx context.Context, userId uuid.UUID) (*dto.User, error) {
	const op = "auth.GetUser"

	log := s.log.With(
		slog.String("op", op),
		slog.String("user_id", userId.String()),
	)

	log.Info("getting user")

	user, err := s.userRepository.GetUser(ctx, userId)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			s.log.Warn("user not found", err)

			return nil, fmt.Errorf("%s: %w", op, err)
		}

		s.log.Error("failed to get user", err)

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user found", slog.String("user_id", user.Username))

	return user, nil

}

func (s *UserService) GetAllFriends(ctx context.Context, userId uuid.UUID) ([]*dto.Friend, error) {
	const op = "auth.GetAllPurchases"

	log := s.log.With(
		slog.String("op", op),
		slog.String("user_id", userId.String()),
	)

	log.Info("getting all purchases")

	purchases, err := s.userRepository.GetAllFriends(ctx, userId)
	if err != nil {
		s.log.Error("failed to get all purchases", err)

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("purchases found", slog.Int("purchases_count", len(purchases)))

	return purchases, nil
}
