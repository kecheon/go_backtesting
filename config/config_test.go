package config_test

import (
	"os"
	"testing"

	"go-backtesting/config"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary test config file
	content := `{
		"filePath": "test_data.csv",
		"vwzPeriod": 10,
		"zscoreThreshold": 2.0,
		"emaPeriod": 20,
		"adxPeriod": 7,
		"adxThreshold": 20.0,
		"boxFilter": {
			"period": 10,
			"minRangePct": 0.01
		},
		"vwzScore": {
			"minStdDev": 1e-5
		}
	}`
	tmpfile, err := os.CreateTemp("", "test_config.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Test loading the config
	cfg, err := config.LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify the loaded config
	if cfg.FilePath != "test_data.csv" {
		t.Errorf("Expected FilePath to be 'test_data.csv', but got '%s'", cfg.FilePath)
	}
	if cfg.VWZPeriod != 10 {
		t.Errorf("Expected VWZPeriod to be 10, but got %d", cfg.VWZPeriod)
	}
	if cfg.ZScoreThreshold != 2.0 {
		t.Errorf("Expected ZScoreThreshold to be 2.0, but got %f", cfg.ZScoreThreshold)
	}
	if cfg.EmaPeriod != 20 {
		t.Errorf("Expected EmaPeriod to be 20, but got %d", cfg.EmaPeriod)
	}
	if cfg.ADXPeriod != 7 {
		t.Errorf("Expected ADXPeriod to be 7, but got %d", cfg.ADXPeriod)
	}
	if cfg.ADXThreshold != 20.0 {
		t.Errorf("Expected ADXThreshold to be 20.0, but got %f", cfg.ADXThreshold)
	}
	if cfg.BoxFilter.Period != 10 {
		t.Errorf("Expected BoxFilter.Period to be 10, but got %d", cfg.BoxFilter.Period)
	}
	if cfg.BoxFilter.MinRangePct != 0.01 {
		t.Errorf("Expected BoxFilter.MinRangePct to be 0.01, but got %f", cfg.BoxFilter.MinRangePct)
	}
	if cfg.VWZScore.MinStdDev != 1e-5 {
		t.Errorf("Expected VWZScore.MinStdDev to be 1e-5, but got %f", cfg.VWZScore.MinStdDev)
	}
}
