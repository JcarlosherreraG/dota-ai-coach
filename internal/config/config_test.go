package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	content := []byte(`{
		"provider": "gemini",
		"model": "gemini-pro",
		"request_interval_seconds": 10,
		"local_gsi_port": 3000
	}`)

	tmpfile, err := os.CreateTemp("", "config_test_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Provider != "gemini" {
		t.Errorf("Expected provider 'openai', got '%s'", cfg.Provider)
	}
	if cfg.GsiPort != 3000 {
		t.Errorf("Expected port 3000, got %d", cfg.GsiPort)
	}
}

func TestLoadConfig_NotFound(t *testing.T) {
	_, err := LoadConfig("non_existent_file.json")
	if err == nil {
		t.Error("Expected error for missing file, got nil")
	}
}

func TestLoadConfig_BadJSON(t *testing.T) {
	tmpfile, _ := os.CreateTemp("", "bad_config_*.json")
	defer os.Remove(tmpfile.Name())

	tmpfile.Write([]byte(`{ "broken": json `))
	tmpfile.Close()

	_, err := LoadConfig(tmpfile.Name())
	if err == nil {
		t.Error("Expected JSON decode error, got nil")
	}
}
