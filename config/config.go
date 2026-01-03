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
	Threshold   float64 `json:"threshold"`
}

type VWZScoreConfig struct {
	MinStdDev float64 `json:"minStdDev"`
}

type VolumeProfileConfig struct {
	LookbackPeriod int     `json:"lookbackPeriod"`
	BinSizePct     float64 `json:"binSizePct"`
	ValueAreaPct   float64 `json:"valueAreaPct"`
	POCProximity   float64 `json:"pocProximity"`
	MinPOCDistance float64 `json:"minPocDistance"`
}

type Config struct {
	FilePath          string              `json:"filePath"`
	VWZPeriod         int             `json:"vwzPeriod"`
	ZScoreThreshold   float64         `json:"zscoreThreshold"`
	EmaPeriod         int             `json:"emaPeriod"`
	BoxFilter         BoxFilterConfig     `json:"boxFilter"`
	VWZScore          VWZScoreConfig      `json:"vwzScore"`
	VolumeCluster     VolumeProfileConfig `json:"volumeCluster"`
	ADXPeriod         int                 `json:"adxPeriod"`
	ADXThreshold      float64         `json:"adxThreshold"`
	AdxUpperThreshold float64         `json:"adxUpperThreshold"`
	TPRate            float64         `json:"TPRate"`
	SLRate            float64         `json:"SLRate"`
	BBWPeriod         int             `json:"bbwPeriod"`
	BBWMultiplier     float64         `json:"bbwMultiplier"`
	BBWThreshold      float64         `json:"bbwThreshold"`
	LongCondition     string          `json:"longCondition"`
	ShortCondition    string          `json:"shortCondition"`
	RunMode           string          `json:"run_mode"`
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
