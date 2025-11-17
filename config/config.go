package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// --- Configuration Structs ---
type BoxFilterConfig struct {
	Period      int     `json:"period"`
	MinRangePct float64 `json:"minRangePct"`
}

type VWZScoreConfig struct {
	MinStdDev float64 `json:"minStdDev"`
}

type Config struct {
	FilePath        string          `json:"filePath"`
	VWZPeriod       int             `json:"vwzPeriod"`
	ZScoreThreshold float64         `json:"zscoreThreshold"`
	EmaPeriod       int             `json:"emaPeriod"`
	BoxFilter       BoxFilterConfig `json:"boxFilter"`
	VWZScore        VWZScoreConfig  `json:"vwzScore"`
	ADXPeriod       int             `json:"adxPeriod"`
	ADXThreshold    float64         `json:"adxThreshold"`
	TPRate          float64         `json:"TPRate"`
	SLRate          float64         `json:"SLRate"`
}

// LoadConfig reads and parses the configuration file.
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	cfg := &Config{}
	err = decoder.Decode(cfg)
	if err != nil {
		return nil, fmt.Errorf("error decoding config file: %w", err)
	}
	return cfg, nil
}
