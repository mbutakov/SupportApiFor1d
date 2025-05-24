package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config содержит конфигурацию приложения
type Config struct {
	Port         string   `json:"port"`
	DatabaseURL  string   `json:"database_url"`
	JWTSecret    string   `json:"jwt_secret"`
	LogFilePath  string   `json:"log_file_path"`
	AllowOrigins []string `json:"allow_origins"`
}

// LoadConfig загружает конфигурацию из файла
func LoadConfig(configPath string) (*Config, error) {
	configFile, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	var config Config
	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	return &Config{
		Port:         "8080",
		DatabaseURL:  "postgres://postgres:postgres@localhost:5432/support_tickets?sslmode=disable",
		JWTSecret:    "your-secret-key",
		LogFilePath:  "logs/app.log",
		AllowOrigins: []string{"*"},
	}
}

// SaveConfig сохраняет конфигурацию в файл
func SaveConfig(config *Config, configPath string) error {
	// Создаем директорию для конфигурационного файла, если ее нет
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configFile, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer configFile.Close()

	jsonEncoder := json.NewEncoder(configFile)
	jsonEncoder.SetIndent("", "  ")
	return jsonEncoder.Encode(config)
}
