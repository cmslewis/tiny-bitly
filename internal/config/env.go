package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Retrieves a time.Duration environment variable in milliseconds. Returns the
// value and an error if the variable is not set or cannot be parsed.
func GetDurationEnv(key string) (time.Duration, error) {
	value, err := GetIntEnv(key)
	if err != nil {
		return 0, err
	}
	return time.Duration(value) * time.Millisecond, nil
}

// Retrieves a time.Duration environment variable in milliseconds, or returns
// the default if not set. The environment variable and default value are
// expected to be in milliseconds.
func GetDurationEnvOrDefault(key string, defaultValue int) time.Duration {
	value := GetIntEnvOrDefault(key, defaultValue)
	return time.Duration(value) * time.Millisecond
}

// Retrieves an integer environment variable. Returns the value and an error if
// the variable is not set or cannot be parsed.
func GetIntEnv(key string) (int, error) {
	valueStr := os.Getenv(key)

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, err
	}

	return value, nil
}

// Retrieves an integer environment variable, or returns the default if not set.
func GetIntEnvOrDefault(key string, defaultValue int) int {
	valueStr := os.Getenv(key)

	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// Retrieves a string environment variable. Returns the value and an error if
// the variable is not set.
func GetStringEnv(key string) (string, error) {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return "", fmt.Errorf("environment variable not set: %s", key)
	}
	return valueStr, nil
}

// Retrieves a string environment variable, or returns the default if not set.
func GetStringEnvOrDefault(key string, defaultValue string) string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	return valueStr
}
