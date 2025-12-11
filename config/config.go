// Файл: config/config.go
package config

import (
	"encoding/json"
	"os"
)

// Config содержит конфигурацию приложения
type Config struct {
	ServiceName string `json:"service_name"`
	ServerPort  int    `json:"server_port"`
	LogLevel    string `json:"log_level"`
	LocalAPIURL string `json:"local_api_url"`
}

// LoadConfig загружает и валидирует конфигурацию из JSON файла
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	// Установка значений по умолчанию, если они не указаны
	if config.ServerPort == 0 {
		config.ServerPort = 8080
	}
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}
	if config.LocalAPIURL == "" {
		config.LocalAPIURL = "http://web:8000"
	}

	return &config, nil
}
