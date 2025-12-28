package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Retrieves a time.Duration environment variable in milliseconds, or returns
// the default if not set. The environment variable and default value are
// expected to be in milliseconds.
func getDurationEnvOrDefault(key string, defaultValue int) time.Duration {
	value := getIntEnvOrDefault(key, defaultValue)
	return time.Duration(value) * time.Millisecond
}

// Retrieves an integer environment variable, or returns the default if not set.
func getIntEnvOrDefault(key string, defaultValue int) int {
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
func getStringEnv(key string) (string, error) {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return "", fmt.Errorf("environment variable not set: %s", key)
	}
	return valueStr, nil
}
