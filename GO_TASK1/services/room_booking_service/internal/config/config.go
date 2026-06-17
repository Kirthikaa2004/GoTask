package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	ProjectRoot string
	BaseUrl     string
	LogLevel    string
	LogDir      string
	LogFileName string
}

func LoadConfig(serviceName, env string) (*Config, error) {
	envFile := fmt.Sprintf("envs/.env.%s", env) 

	if err := godotenv.Load(envFile); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("ENV file isn't present, continuing without it:", envFile)
		} else {
			return nil, fmt.Errorf("failed to load %s: %w", envFile, err)
		}
	}

	config := &Config{
		Port:        getEnv("PORT", ""),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		ProjectRoot: getEnv("PROJECT_ROOT", ""),
		BaseUrl:     getEnv("BASE_URL", ""),
		LogLevel:    getEnv("BASE_URL","info"),
	    LogDir:      getEnv("Log_Dir","./logs"),
	    LogFileName: getEnv("Log_File_Name","room-booking.log"),
	}

	if config.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is required")
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
