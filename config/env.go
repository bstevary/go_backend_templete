package config

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Env stores all configuration of the application.
type Env struct {
	Environment          string        `mapstructure:"ENVIRONMENT"`
	DBSource             string        `mapstructure:"DB_SOURCE"`
	RedisURL             string        `mapstructure:"REDIS_URL"`
	ServerAddress        string        `mapstructure:"SERVER_ADDRESS"`
	WorkerID             int64         `mapstructure:"WORKER_ID"`
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	RabbitMQURL          string        `mapstructure:"RABBITMQ_URL"`
	Domain               string        `mapstructure:"DOMAIN"`
	AllowedOrigins       []string      `mapstructure:"ALLOWED_ORIGINS"`
	RedisConfig          *RedisConfig  `mapstructure:"-"`
	AwsRegion            string        `mapstructure:"AWS_REGION"`
	S3BucketName         string        `mapstructure:"S3_BUCKET_NAME"`
	FileStoragePath      string        `mapstructure:"FILE_STORAGE_PATH"`
	LogFilePath          string        `mapstructure:"LOG_FILE_PATH"`
	EmailSenderName      string        `mapstructure:"EMAIL_SENDER_NAME"`
	EmailSenderAddress   string        `mapstructure:"EMAIL_SENDER_ADDRESS"`
	EmailSenderPassword  string        `mapstructure:"EMAIL_SENDER_PASSWORD"`
}

// LoadEnv reads configuration from file or environment variables.
func LoadEnv(path string) (env *Env, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigType("env")
	viper.SetConfigName(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Info().Err(err).Msg(".env not found, proceeding without it")
		} else {
			return &Env{}, err
		}
	}

	if err := viper.Unmarshal(&env); err != nil {
		return &Env{}, err
	}

	// Parse Redis URL after unmarshaling
	if env.RedisURL != "" {
		redisConfig, err := parseRedisURL(env.RedisURL)
		if err != nil {
			return &Env{}, fmt.Errorf("failed to parse REDIS_URL: %w", err)
		}
		env.RedisConfig = redisConfig
	}

	// ✅ Enforce presence of WorkerID
	if env.WorkerID < 1 {
		return &Env{}, fmt.Errorf("WORKER_ID must be provided and greater than 0")
	}

	// ✅ Normalize WorkerID to [1, 1023]
	env.WorkerID = normalizeWorkerID(env.WorkerID)

	log.Info().Int64("worker_id", env.WorkerID).Msg("Worker ID normalized and validated")

	return env, nil
}

// normalizeWorkerID ensures any number (up to 65535 or more) maps into [1,1023].
func normalizeWorkerID(id int64) int64 {
	const maxWorkerID = 1023
	const privilegedLimit = 1024

	if id < privilegedLimit {
		return id % maxWorkerID
	}
	mapped := (id - privilegedLimit) % maxWorkerID
	if mapped == 0 {
		mapped = maxWorkerID
	}
	return mapped
}
