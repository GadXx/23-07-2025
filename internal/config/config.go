package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port            string
	AllowedExts     []string
	MaxFilesPerTask int
	MaxActiveTasks  int
}

func GetEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return i
}

func DefaultConfig() *Config {
	return &Config{
		Port:            os.Getenv("SERVER_PORT"),
		AllowedExts:     []string{".pdf", ".jpeg"},
		MaxFilesPerTask: GetEnvInt("MAX_SESSIONS", 10),
		MaxActiveTasks:  GetEnvInt("MAX_ARCHIVE_SIZE", 10),
	}
}
