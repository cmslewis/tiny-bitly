package config

import (
	"log/slog"
	"time"
	"tiny-bitly/internal/apperrors"
)

type Config struct {
	// Server
	APIPort     int
	APIHostname string
	LogLevel    slog.Leveler

	// Limits
	MaxAliasLength          int
	MaxRequestSizeBytes     int
	MaxTriesCreateShortCode int
	MaxURLLength            int
	ShortCodeLength         int

	// Database
	PostgresPort     int
	PostgresDB       string
	PostgresUser     string
	PostgresPassword string

	// Redis
	RedisHost string
	RedisPort int

	// Rate Limiting
	RateLimitRequestsPerSecond int
	RateLimitBurst             int

	// Timeouts
	IdleTimeout     time.Duration
	ReadTimeout     time.Duration
	RequestTimeout  time.Duration
	ShortCodeTTL    time.Duration
	ShutdownTimeout time.Duration
	WriteTimeout    time.Duration
}

// Loads environment variables and returns them into a central config object.
// Returns an error if critical config is not defined in env files.
func LoadConfig() (*Config, error) {
	hostname, err := getStringEnv("API_HOSTNAME")
	if err != nil {
		return nil, apperrors.ErrConfigurationMissing
	}

	port := getIntEnvOrDefault("API_PORT", defaultAPIPort)
	logLevel := getStringEnvOrDefault("LOG_LEVEL", defaultLogLevel)

	rateLimitRPS := getIntEnvOrDefault("RATE_LIMIT_REQUESTS_PER_SECOND", defaultRateLimitRequestsPerSecond)
	rateLimitBurst := getIntEnvOrDefault("RATE_LIMIT_BURST", defaultRateLimitBurst)

	maxAliasLength := getIntEnvOrDefault("MAX_ALIAS_LENGTH", defaultMaxAliasLength)
	maxRequestSizeBytes := getIntEnvOrDefault("MAX_REQUEST_SIZE_BYTES", defaultMaxRequestSizeBytes)
	maxTries := getIntEnvOrDefault("MAX_TRIES_CREATE_SHORT_CODE", defaultMaxTriesCreateShortCode)
	maxURLLength := getIntEnvOrDefault("MAX_URL_LENGTH", defaultMaxUrlLength)
	shortCodeLength := getIntEnvOrDefault("SHORT_CODE_LENGTH", defaultShortCodeLength)
	shortCodeTTL := getDurationEnvOrDefault("SHORT_CODE_TTL_MILLIS", defaultShortCodeTtlMillis)

	postgresPort := getIntEnvOrDefault("POSTGRES_PORT", defaultPostgresPort)
	postgresDB := getStringEnvOrDefault("POSTGRES_DB", defaultPostgresDB)
	postgresUser := getStringEnvOrDefault("POSTGRES_USER", defaultPostgresUser)
	postgresPassword := getStringEnvOrDefault("POSTGRES_PASSWORD", defaultPostgresPassword)

	redisHost := getStringEnvOrDefault("REDIS_HOST", defaultRedisHost)
	redisPort := getIntEnvOrDefault("REDIS_PORT", defaultRedisPort)

	idleTimeout := getDurationEnvOrDefault("TIMEOUT_IDLE_MILLIS", defaultTimeoutIdleMillis)
	requestTimeout := getDurationEnvOrDefault("TIMEOUT_REQUEST_MILLIS", defaultTimeoutRequestMillis)
	readTimeout := getDurationEnvOrDefault("TIMEOUT_READ_MILLIS", defaultTimeoutReadMillis)
	shutdownTimeout := getDurationEnvOrDefault("TIMEOUT_SHUTDOWN_MILLIS", defaultTimeoutShutdownMillis)
	writeTimeout := getDurationEnvOrDefault("TIMEOUT_WRITE_MILLIS", defaultTimeoutWriteMillis)

	return &Config{
		APIPort:     port,
		APIHostname: hostname,
		LogLevel:    getLogLevelTyped(logLevel),

		RateLimitRequestsPerSecond: rateLimitRPS,
		RateLimitBurst:             rateLimitBurst,

		MaxAliasLength:          maxAliasLength,
		MaxRequestSizeBytes:     maxRequestSizeBytes,
		MaxTriesCreateShortCode: maxTries,
		MaxURLLength:            maxURLLength,
		ShortCodeLength:         shortCodeLength,

		PostgresPort:     postgresPort,
		PostgresDB:       postgresDB,
		PostgresUser:     postgresUser,
		PostgresPassword: postgresPassword,

		RedisHost: redisHost,
		RedisPort: redisPort,

		IdleTimeout:     idleTimeout,
		ReadTimeout:     readTimeout,
		RequestTimeout:  requestTimeout,
		ShortCodeTTL:    shortCodeTTL,
		ShutdownTimeout: shutdownTimeout,
		WriteTimeout:    writeTimeout,
	}, nil
}

func getLogLevelTyped(logLevel string) slog.Level {
	if logLevel == "debug" {
		return slog.LevelDebug
	}
	return slog.LevelInfo
}
