package config

import (
	_ "embed"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

//go:embed config.yaml
var defaultConfigYAML []byte

type Config struct {
	ServerPort string `yaml:"server_port"`
}

var cfg Config

func Load() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	exeDir := filepath.Dir(exePath)
	configPath := filepath.Join(exeDir, "config.yaml")

	// Если файл конфигурации не существует, создаем его из встроенного шаблона
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := os.WriteFile(configPath, defaultConfigYAML, 0644); err != nil {
			return err
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return err
	}

	cfg = c
	return nil
}

func Get() Config {
	return cfg
}

