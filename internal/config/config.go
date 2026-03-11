// Package config управляет конфигурацией приложения (YAML).
// Хранение API-ключа в ~/.config/ai-launcher/config.yaml с правами 0600 (FR-105, FR-106).
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config — основная структура конфигурации.
type Config struct {
	APIKey       string            `yaml:"api_key,omitempty"`
	Tools        []ToolEntry       `yaml:"tools,omitempty"`
	Telemetry    TelemetryConfig   `yaml:"telemetry,omitempty"`
	MCPRegistry  string            `yaml:"mcp_registry,omitempty"` // URL GitLab npm registry для MCP
}

// ToolEntry — запись об инструменте (редактируемые поля FR-401).
type ToolEntry struct {
	ID       string            `yaml:"id"`
	Name     string            `yaml:"name"`
	Command  string            `yaml:"command"`
	Model    string            `yaml:"model"`
	Env      map[string]string `yaml:"env,omitempty"`
	Enabled  bool              `yaml:"enabled"`
	Favorite bool              `yaml:"favorite"`
}

// TelemetryConfig — настройки телеметрии (FR-705).
type TelemetryConfig struct {
	Enabled      bool   `yaml:"enabled"`
	OTLPEndpoint string `yaml:"otlp_endpoint,omitempty"`
}

// ConfigPath возвращает путь к файлу конфигурации.
func ConfigPath() (string, error) {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "ai-launcher", "config.yaml"), nil
}

// Load загружает конфигурацию из файла.
func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// Save сохраняет конфигурацию. Создаёт каталог и выставляет права 0600 на файл.
func Save(c *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
