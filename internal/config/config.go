package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	PageTitle       string   `json:"pageTitle"`
	PageTitleSuffix string   `json:"pageTitleSuffix"`
	Locale          string   `json:"locale"`
	BaseURL         string   `json:"baseURL"`
	IgnorePatterns  []string `json:"ignorePatterns"`
	PublishMode     string   `json:"publishMode"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
