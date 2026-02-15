package config

import (
	"os"
	"testing"
)

func TestGetGrammarForExtension(t *testing.T) {
	c := &Config{}
	c.Grammar.Mapping = map[string]string{
		".py": "python",
		".ts": "typescript",
	}

	tests := []struct {
		ext      string
		expected string
	}{
		{".py", "python"},
		{".ts", "typescript"},
		{".go", "go"},         // default fallback
		{".random", "random"}, // default fallback
	}

	for _, tt := range tests {
		if got := c.GetGrammarForExtension(tt.ext); got != tt.expected {
			t.Errorf("GetGrammarForExtension(%q) = %q, want %q", tt.ext, got, tt.expected)
		}
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	// Ensure we don't pick up actual user config during tests
	if err := os.Setenv("COGENI_CONFIG", "non-existent"); err != nil {
		t.Fatalf("Failed to set env: %v", err)
	}
	defer func() { _ = os.Unsetenv("COGENI_CONFIG") }()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Grammar.Mapping[".py"] != "python" {
		t.Errorf("Expected default mapping .py -> python, got %q", cfg.Grammar.Mapping[".py"])
	}

	if cfg.Grammar.Mapping[".ts"] != "typescript" {
		t.Errorf("Expected default mapping .ts -> typescript, got %q", cfg.Grammar.Mapping[".ts"])
	}
}
