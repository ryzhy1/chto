package redis

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

type Storage struct {
	db         *redis.Client
	refreshTTL time.Duration
}

// InitRedis инициализирует клиент Redis.
func InitRedis(connStr, redisUsername, redisPassword, redisDbNumber string, maxRetries int, redisTimeout, refreshTTL time.Duration) (*Storage, error) {
	dbNumber, err := strconv.Atoi(redisDbNumber)
	if err != nil {
		return nil, err
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:         connStr,
		Username:     redisUsername,
		Password:     redisPassword,
		DB:           dbNumber,
		MaxRetries:   maxRetries,
		ReadTimeout:  redisTimeout,
		WriteTimeout: redisTimeout,
	})
	return &Storage{db: redisClient, refreshTTL: refreshTTL}, nil
}

var ctx = context.Background()

func (s *Storage) StoreRefreshToken(userID string) (string, error) {
	refreshToken, err := uuidGenerator()
	if err != nil {
		return "", err
	}

	tokenData := map[string]interface{}{
		"user_id": userID,
		"issued":  time.Now().Unix(),
	}

	ttl := 7 * 24 * time.Hour

	err = s.db.HSet(ctx, refreshToken, tokenData).Err()
	if err != nil {
		return "", err
	}

	err = s.db.Expire(ctx, refreshToken, ttl).Err()
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func VerifyRefreshToken(redisClient *redis.Client, refreshToken string) (map[string]string, error) {
	tokenData, err := redisClient.HGetAll(ctx, refreshToken).Result()
	if err != nil || len(tokenData) == 0 {
		return nil, fmt.Errorf("invalid or expired refresh token")
	}

	return tokenData, nil
}

func RevokeRefreshToken(redisClient *redis.Client, refreshToken string) error {
	return redisClient.Del(ctx, refreshToken).Err()
}

func (s *Storage) CloseConnection() error {
	err := s.db.Close()
	if err != nil {
		return err
	}
	return nil
}

func uuidGenerator() (string, error) {
	uid, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	uidString := uid.String()

	return uidString, nil
}
