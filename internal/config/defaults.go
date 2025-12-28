package config

import (
	"fmt"
	"time"
)

var defaultAPIPort int = 8080
var defaultAPIHostname string = fmt.Sprintf("http://localhost:%d", defaultAPIPort)
var defaultLogLevel string = "info"
var defaultMaxTriesCreateShortCode int = 10
var defaultMaxUrlLength int = 1000
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
		APIPort:                 defaultAPIPort,
		APIHostname:             defaultAPIHostname,
		LogLevel:                getLogLevelTyped(defaultLogLevel),
		MaxTriesCreateShortCode: defaultMaxTriesCreateShortCode,
		MaxURLLength:            defaultMaxUrlLength,
		ShortCodeLength:         defaultShortCodeLength,
		ShortCodeTTL:            time.Duration(defaultShortCodeTtlMillis) * time.Millisecond,
		IdleTimeout:             time.Duration(defaultTimeoutIdleMillis) * time.Millisecond,
		ReadTimeout:             time.Duration(defaultTimeoutReadMillis) * time.Millisecond,
		RequestTimeout:          time.Duration(defaultTimeoutRequestMillis) * time.Millisecond,
		ShutdownTimeout:         time.Duration(defaultTimeoutShutdownMillis) * time.Millisecond,
		WriteTimeout:            time.Duration(defaultTimeoutWriteMillis) * time.Millisecond,
	}
}
