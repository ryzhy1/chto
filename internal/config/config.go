package config

import (
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"time"
)

type ServerConfig struct {
	Env         string        `env:"ENV,required"` // local, dev, prod
	AuthAddress string        `env:"AUTH_ADDRESS,required"`
	UserAddress string        `env:"USER_ADDRESS,required"`
	AuthTimeout time.Duration `env:"AUTH_TIMEOUT" envDefault:"5s"`
	UserTimeout time.Duration `env:"USER_TIMEOUT" envDefault:"5s"`
}

type DatabaseConfig struct {
	PostgresConn string `env:"POSTGRES_CONN,required"`
}

type JWTConfig struct {
	Secret                  string        `env:"JWT_SECRET,required"`
	AccessExpirationMinutes time.Duration `env:"ACCESS_EXPIRATION_MINUTES" envDefault:"15"`
	RefreshExpirationDays   time.Duration `env:"REFRESH_EXPIRATION_DAYS" envDefault:"7"`
}

type RedisConfig struct {
	RedisConn     string        `env:"REDIS_CONN,required"`
	RedisUsername string        `env:"REDIS_USERNAME,required"`
	RedisPassword string        `env:"REDIS_PASSWORD,required"`
	RedisDbNumber int           `env:"REDIS_DB_NUMBER,required"`
	MaxRetries    int           `env:"MAX_RETRIES" envDefault:"3"`
	Timeout       time.Duration `env:"TIMEOUT" envDefault:"5s"`
}

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
}

const (
	local = ".env.local"
	dev   = ".env.dev"
	prod  = ".env.prod"
)

func MustLoad() *Config {
	if err := godotenv.Load(local); err != nil {
		panic(err)
	}

	authTimeoutStr := os.Getenv("AUTH_TIMEOUT")
	authTimeout, err := time.ParseDuration(authTimeoutStr)
	if err != nil {
		panic("Invalid TIMEOUT format: " + err.Error())
	}

	userTimeoutStr := os.Getenv("USER_TIMEOUT")
	userTimeout, err := time.ParseDuration(userTimeoutStr)
	if err != nil {
		panic("Invalid TIMEOUT format: " + err.Error())
	}

	accessExpStr := os.Getenv("AUTH_ACCESS_EXPIRATION_MINUTES")
	accessExp, err := time.ParseDuration(accessExpStr)
	if err != nil {
		panic("Invalid AUTH_ACCESS_EXPIRATION_MINUTES format: " + err.Error())
	}

	refreshExpStr := os.Getenv("AUTH_REFRESH_EXPIRATION_DAYS")
	refreshExp, err := time.ParseDuration(refreshExpStr)
	if err != nil {
		panic("Invalid AUTH_REFRESH_EXPIRATION_DAYS format: " + err.Error())
	}

	redisDbNumberStr := os.Getenv("REDIS_DB_NUMBER")
	redisDbNumber, err := strconv.Atoi(redisDbNumberStr)
	if err != nil {
		panic("Invalid REDIS_DB_NUMBER format: " + err.Error())
	}

	redisMaxRetriesStr := os.Getenv("REDIS_MAX_RETRIES")
	redisMaxRetries, err := strconv.Atoi(redisMaxRetriesStr)
	if err != nil {
		panic("Invalid REDIS_MAX_RETRIES format: " + err.Error())
	}

	redisTimeoutStr := os.Getenv("REDIS_TIMEOUT")
	redisTimeout, err := time.ParseDuration(redisTimeoutStr)
	if err != nil {
		panic("Invalid REDIS_TIMEOUT format: " + err.Error())
	}

	return &Config{
		Server: ServerConfig{
			Env:         os.Getenv("ENV"),
			AuthAddress: os.Getenv("AUTH_ADDRESS"),
			UserAddress: os.Getenv("USER_ADDRESS"),
			AuthTimeout: authTimeout,
			UserTimeout: userTimeout,
		},
		Database: DatabaseConfig{
			PostgresConn: os.Getenv("POSTGRES_CONN"),
		},
		Redis: RedisConfig{
			RedisConn:     os.Getenv("REDIS_STORAGE_PATH"),
			RedisUsername: os.Getenv("REDIS_USERNAME"),
			RedisPassword: os.Getenv("REDIS_PASSWORD"),
			RedisDbNumber: redisDbNumber,
			MaxRetries:    redisMaxRetries,
			Timeout:       redisTimeout,
		},
		JWT: JWTConfig{
			Secret:                  os.Getenv("JWT_SECRET"),
			AccessExpirationMinutes: accessExp,
			RefreshExpirationDays:   refreshExp,
		},
	}
}
