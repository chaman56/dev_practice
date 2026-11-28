package env

import (
	"os"
	"strconv"
	"time"
)

func GetEnv(key, defaultValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultValue
}

func GetEnvAsInt(key string, defaultValue int) int {
	valStr, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	var val int
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue
	}
	return val
}

func GetEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if valStr, ok := os.LookupEnv(key); ok {
		val, err := time.ParseDuration(valStr)
		if err == nil {
			return val
		}
	}
	return defaultValue
}
