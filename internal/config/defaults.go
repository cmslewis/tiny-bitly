package config

import (
	"fmt"
	"time"
)

var defaultAPIPort int = 8080
var defaultAPIHostname string = fmt.Sprintf("http://localhost:%d", defaultAPIPort)
var defaultLogLevel string = "info"
var defaultMaxAliasLength int = 30
var defaultMaxRequestSizeBytes int = 1048576 // 1 MB, reasonable for a URL shortening service
var defaultMaxTriesCreateShortCode int = 10
var defaultMaxUrlLength int = 1000
var defaultPostgresPort int = 5434
var defaultPostgresDB string = ""
var defaultPostgresUser string = ""
var defaultPostgresPassword string = ""

var defaultRateLimitBurst int = 20
var defaultRateLimitRequestsPerSecond int = 10
var defaultShortCodeLength int = 6
var defaultShortCodeTtlMillis int = 30000
var defaultTimeoutIdleMillis int = 60000
var defaultTimeoutReadMillis int = 30000
var defaultTimeoutRequestMillis int = 30000
var defaultTimeoutShutdownMillis int = 30000
var defaultTimeoutWriteMillis int = 30000

// Returns a config object with sensible defaults in place for each key.
func GetDefaultConfig() Config {
	return Config{
		APIPort:                    defaultAPIPort,
		APIHostname:                defaultAPIHostname,
		LogLevel:                   getLogLevelTyped(defaultLogLevel),
		RateLimitRequestsPerSecond: defaultRateLimitRequestsPerSecond,
		RateLimitBurst:             defaultRateLimitBurst,
		MaxAliasLength:             defaultMaxAliasLength,
		MaxRequestSizeBytes:        defaultMaxRequestSizeBytes,
		MaxTriesCreateShortCode:    defaultMaxTriesCreateShortCode,
		MaxURLLength:               defaultMaxUrlLength,
		PostgresPort:               defaultPostgresPort,
		PostgresDB:                 defaultPostgresDB,
		PostgresUser:               defaultPostgresUser,
		PostgresPassword:           defaultPostgresPassword,
		ShortCodeLength:            defaultShortCodeLength,
		ShortCodeTTL:               time.Duration(defaultShortCodeTtlMillis) * time.Millisecond,
		IdleTimeout:                time.Duration(defaultTimeoutIdleMillis) * time.Millisecond,
		ReadTimeout:                time.Duration(defaultTimeoutReadMillis) * time.Millisecond,
		RequestTimeout:             time.Duration(defaultTimeoutRequestMillis) * time.Millisecond,
		ShutdownTimeout:            time.Duration(defaultTimeoutShutdownMillis) * time.Millisecond,
		WriteTimeout:               time.Duration(defaultTimeoutWriteMillis) * time.Millisecond,
	}
}
