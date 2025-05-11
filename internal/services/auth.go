package services

import (
	"boton-back/internal/domain/models"
	_ "boton-back/internal/lib/jwt"
	"boton-back/internal/repository"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"regexp"
	"time"
)

var (
	ErrEmptyField       = errors.New("all fields must be filled")
	ErrInvalidEmail     = errors.New("email is invalid")
	ErrLoginTooShort    = errors.New("login must be at least 3 characters")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
)

type AuthService struct {
	log            *slog.Logger
	redisDB        RedisClient
	authRepository AuthRepository
	tokenTTL       time.Duration
	jwtGenerator   JwtGenerator
}

type JwtGenerator interface {
	GeneratePair(id uuid.UUID) (accessToken string, refreshToken string, err error)
	ParseToken(tokenString string) (string, error)
}

type AuthRepository interface {
	SaveUser(ctx context.Context, login, email string, password []byte) error
	LoginUser(ctx context.Context, inputType, input string) (*models.User, error)
	CheckUsernameIsAvailable(ctx context.Context, login string) (bool, error)
	CheckEmailIsAvailable(ctx context.Context, email string) (bool, error)
	CheckUserByEmail(ctx context.Context, userId, email string) error
	CheckUserByPassword(ctx context.Context, userId, password string) (string, error)
	UpdateEmail(ctx context.Context, userId, email string) error
	UpdatePassword(ctx context.Context, userId, password string) error
}

type RedisClient interface {
	StoreRefreshToken(userID string) (string, error)
	CloseConnection() error
}

var (
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrUserAlreadyExists    = errors.New("user already exists")
	ErrUserNotFound         = errors.New("user not found")
	ErrEmailAlreadyTaken    = errors.New("this email already taken")
	ErrUsernameAlreadyTaken = errors.New("this username already taken")
)

func NewAuthService(log *slog.Logger, jwtGenerator JwtGenerator, authRepository AuthRepository, redisDB RedisClient) *AuthService {
	return &AuthService{
		log:            log,
		jwtGenerator:   jwtGenerator,
		redisDB:        redisDB,
		authRepository: authRepository,
	}
}

func (s *AuthService) Register(ctx context.Context, login, email, password string) error {
	const op = "auth.Register"

	log := s.log.With(slog.String("op", op), slog.String("email", email))

	if err := checkRegister(login, email, password); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.checkContext(ctx, op); err != nil {
		return err
	}

	usernameAvailable, err := s.authRepository.CheckUsernameIsAvailable(ctx, login)
	if err != nil {
		log.Error("failed to check username availability", err)
		return fmt.Errorf("%s: failed to check username: %w", op, err)
	}

	if !usernameAvailable {
		return fmt.Errorf("%s: %w", op, ErrUsernameAlreadyTaken)
	}

	if err := s.checkContext(ctx, op); err != nil {
		return err
	}

	emailAvailable, err := s.authRepository.CheckEmailIsAvailable(ctx, email)
	if err != nil {
		log.Error("failed to check email availability", err)
		return fmt.Errorf("%s: failed to check email: %w", op, err)
	}

	if !emailAvailable {
		return fmt.Errorf("%s: %w", op, ErrEmailAlreadyTaken)
	}

	log.Info("registering new user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to hash password", err)
		return fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	log.Info("password hash created")

	if err := s.authRepository.SaveUser(ctx, login, email, passHash); err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			log.Warn("user already exists", err)
			return fmt.Errorf("%s: %w", op, err)
		}
		log.Error("failed to save user", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered", "id")
	return nil
}

func (s *AuthService) Login(ctx context.Context, input, password string) (string, string, error) {
	const op = "auth.Login"

	log := s.log.With(
		slog.String("op", op),
		slog.String("input", input),
	)

	if err := checkLogin(input, password); err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("logging in")

	inputType := identifyLoginInputType(input)

	user, err := s.authRepository.LoginUser(ctx, inputType, input)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			s.log.Warn("user not found", err)

			return "", "", fmt.Errorf("%s: %w", op, err)
		}

		s.log.Error("failed to get user", err)

		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	if err = bcrypt.CompareHashAndPassword(user.Password, []byte(password)); err != nil {
		s.log.Info("invalid credentials", err)

		return "", "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	accessToken, refreshToken, err := s.jwtGenerator.GeneratePair(user.ID)
	if err != nil {
		s.log.Error("failed to generate access token", err)
		return "", "", fmt.Errorf("%s: %w", op, err)
	}
	return accessToken, refreshToken, nil
}

func (s *AuthService) UpdateUserEmail(ctx context.Context, userId, oldEmail, newEmail string) (string, error) {
	const op = "auth.UpdateUserEmail"

	log := s.log.With(
		slog.String("op", op),
		slog.String("userId", userId),
	)

	log.Info("getting user email")

	if !correctEmailChecker(oldEmail) {
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	if !correctEmailChecker(newEmail) {
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	err := s.authRepository.CheckUserByEmail(ctx, userId, oldEmail)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			s.log.Warn("user not found", err)

			return "", fmt.Errorf("%s: %w", op, err)
		}

		s.log.Error("failed to get user", err)

		return "", fmt.Errorf("%s: %w", op, err)
	}

	err = s.authRepository.UpdateEmail(ctx, userId, newEmail)
	if err != nil {
		s.log.Error("failed to update user email", err)

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return "email updated successfully", nil
}

func (s *AuthService) UpdateUserPassword(ctx context.Context, userId, oldPassword, newPassword string) (string, error) {
	const op = "auth.UpdateUserPassword"

	log := s.log.With(
		slog.String("op", op),
		slog.String("userId", userId),
	)

	log.Info("checking user credentials")

	if len(oldPassword) < 8 || len(newPassword) < 8 {
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	if oldPassword == newPassword {
		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	log.Info("picking user password from database")

	password, err := s.authRepository.CheckUserByPassword(ctx, userId, oldPassword)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			s.log.Warn("user not found", err)

			return "", fmt.Errorf("%s: %w", op, err)
		}

		s.log.Error("failed to get user", err)

		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("comparing users password")

	err = bcrypt.CompareHashAndPassword([]byte(password), []byte(oldPassword))
	if err != nil {
		s.log.Info("invalid credentials", err)

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	log.Info("hashing new password")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("failed to hash password", err)

		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("updating user password")

	err = s.authRepository.UpdatePassword(ctx, userId, string(hashedPassword))
	if err != nil {
		s.log.Error("failed to update user password", err)

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return "password updated successfully", nil
}

func (s *AuthService) checkContext(ctx context.Context, op string) error {
	if ctx.Err() != nil {
		return fmt.Errorf("%s: context canceled: %w", op, ctx.Err())
	}
	return nil
}

func checkRegister(login, email, password string) error {
	if login == "" || email == "" || password == "" {
		return ErrEmptyField
	}

	if !correctEmailChecker(email) {
		return ErrInvalidEmail
	}

	if len(login) < 3 {
		return fmt.Errorf("%w: minimum 3 characters required", ErrLoginTooShort)
	}

	if len(password) < 8 {
		return fmt.Errorf("%w: minimum 8 characters required", ErrPasswordTooShort)
	}

	return nil
}

func correctEmailChecker(email string) bool {
	const emailPattern = `^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`
	emailRegex := regexp.MustCompile(emailPattern)

	if emailRegex.MatchString(email) {
		return true
	}

	return false
}

func checkLogin(login, password string) error {
	if login == "" || password == "" {
		return ErrEmptyField
	}

	if len(login) < 3 {
		return fmt.Errorf("%w: minimum 3 characters required", ErrLoginTooShort)
	}

	if len(password) < 8 {
		return fmt.Errorf("%w: minimum 8 characters required", ErrPasswordTooShort)
	}

	return nil
}

func identifyLoginInputType(input string) string {
	if correctEmailChecker(input) {
		return "email"
	}
	return "username"
}
