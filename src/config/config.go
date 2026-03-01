// Package config provides the configuration infrastructure for cogeni.
// It handles loading settings from YAML files, environment variables, and defaults,
// managing grammar locations, file extension mappings, and custom grammar build sources.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// appName is used for configuration file names and directory paths.
	appName = "cogeni"
)

// GrammarSource defines how a specific Tree-sitter grammar should be obtained and built.
// This is used when the default automatic download from the Tree-sitter GitHub organization
// is insufficient or when a custom grammar repository is required.
type GrammarSource struct {
	// URL is the Git repository URL containing the grammar source code.
	URL string `yaml:"url" json:"url"`
	// Branch is the specific Git branch to use (e.g., "master", "main", "develop").
	Branch string `yaml:"branch" json:"branch"`
	// BuildCmd is an optional shell command to execute for building the grammar.
	// If omitted, the manager will attempt a standard C/C++ compilation.
	BuildCmd string `yaml:"build_cmd" json:"build_cmd"`
	// Artifact is the name or relative path of the resulting shared library
	// (e.g., "python.so", "build/sql.dylib").
	Artifact string `yaml:"artifact" json:"artifact"`
}

// Config is the root configuration object for the cogeni application.
type Config struct {
	// Grammar contains all settings related to Tree-sitter grammars.
	Grammar struct {
		// Location is the filesystem path where grammar shared libraries are cached.
		Location string `yaml:"location" json:"location"`
		// Mapping maps file extensions (e.g., ".py") to grammar names (e.g., "python").
		Mapping map[string]string `yaml:"mapping" json:"mapping"`
		// Sources provides custom build instructions for specific grammar names.
		Sources map[string]GrammarSource `yaml:"sources" json:"sources"`
	} `yaml:"grammar" json:"grammar"`
}

// LoadConfig initializes the application configuration.
// It searches for "config.yaml" in standard OS-specific configuration directories
// (e.g., ~/.config/cogeni/ on Linux) and applies defaults and environment variable overrides.
// Environment variables follow the pattern COGENI_PATH_TO_KEY (e.g., COGENI_GRAMMAR_LOCATION).
func LoadConfig() (*Config, error) {
	cfg := &Config{}

	// Determine config file path based on OS and XDG standards
	configPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to determine config path: %w", err)
	}

	configFilePath := filepath.Join(configPath, "config.yaml")
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found, defaults will be used.
		fmt.Fprintf(os.Stderr, "Config file not found in %s, using defaults.\n", configPath)
	} else {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	// Apply defaults
	defaultGrammarLocation, err := getDefaultGrammarLocation()
	if err != nil {
		return nil, fmt.Errorf("failed to determine default grammar location: %w", err)
	}

	if cfg.Grammar.Location == "" {
		cfg.Grammar.Location = defaultGrammarLocation
	}

	if cfg.Grammar.Mapping == nil {
		cfg.Grammar.Mapping = make(map[string]string)
	}
	defaultMapping := map[string]string{
		".py":   "python",
		".ts":   "typescript",
		".js":   "javascript",
		".go":   "go",
		".lua":  "lua",
		".sql":  "sql",
		".sh":   "bash",
		".bash": "bash",
		".json": "json",
		".yaml": "yaml",
		".yml":  "yaml",
		".toml": "toml",
	}
	for k, v := range defaultMapping {
		if _, exists := cfg.Grammar.Mapping[k]; !exists {
			cfg.Grammar.Mapping[k] = v
		}
	}

	// Environment variable overrides
	if envLoc := os.Getenv("COGENI_GRAMMAR_LOCATION"); envLoc != "" {
		cfg.Grammar.Location = envLoc
	}

	// Iterate over environment variables and apply overrides
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, val := parts[0], parts[1]

		// Support COGENI_GRAMMAR_MAPPING_EXT=LANG
		if strings.HasPrefix(key, "COGENI_GRAMMAR_MAPPING_") {
			ext := strings.TrimPrefix(key, "COGENI_GRAMMAR_MAPPING_")
			ext = "." + strings.ToLower(ext)
			cfg.Grammar.Mapping[ext] = val
		}

		// Support COGENI_GRAMMAR_SOURCES_NAME_FIELD=VALUE
		if strings.HasPrefix(key, "COGENI_GRAMMAR_SOURCES_") {
			rest := strings.TrimPrefix(key, "COGENI_GRAMMAR_SOURCES_")
			// We expect NAME_FIELD, e.g. SQL_URL
			parts := strings.SplitN(rest, "_", 2)
			if len(parts) != 2 {
				continue
			}
			name := strings.ToLower(parts[0])
			field := strings.ToLower(parts[1])

			if cfg.Grammar.Sources == nil {
				cfg.Grammar.Sources = make(map[string]GrammarSource)
			}

			source := cfg.Grammar.Sources[name]
			switch field {
			case "url":
				source.URL = val
			case "branch":
				source.Branch = val
			case "build_cmd":
				source.BuildCmd = val
			case "artifact":
				source.Artifact = val
			}
			cfg.Grammar.Sources[name] = source
		}
	}

	return cfg, nil
}

// getConfigPath determines the configuration directory based on OS.
func getConfigPath() (string, error) {
	var configDir string
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "windows":
		configDir = os.Getenv("APPDATA")
		if configDir == "" {
			return "", fmt.Errorf("appdata environment variable not set")
		}
	case "darwin":
		configDir = filepath.Join(home, "Library", "Application Support")
	case "linux":
		// XDG Base Directory Specification
		configDir = os.Getenv("XDG_CONFIG_HOME")
		if configDir == "" {
			configDir = filepath.Join(home, ".config")
		}
	default:
		configDir = filepath.Join(home, ".config") // Fallback for other Unix-like systems
	}

	return filepath.Join(configDir, appName), nil
}

// getDefaultGrammarLocation determines the default grammar storage directory based on OS.
func getDefaultGrammarLocation() (string, error) {
	var dataDir string
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "windows":
		dataDir = os.Getenv("LOCALAPPDATA")
		if dataDir == "" {
			dataDir = os.Getenv("APPDATA")
		}
		if dataDir == "" {
			return "", fmt.Errorf("localappdata or appdata environment variable not set")
		}
	case "darwin":
		dataDir = filepath.Join(home, "Library", "Application Support")
	case "linux":
		// XDG Base Directory Specification
		dataDir = os.Getenv("XDG_DATA_HOME")
		if dataDir == "" {
			dataDir = filepath.Join(home, ".local", "share")
		}
	default:
		dataDir = filepath.Join(home, ".local", "share") // Fallback for other Unix-like systems
	}

	return filepath.Join(dataDir, appName, "grammars"), nil
}

// GetGrammarForExtension returns the tree-sitter grammar name for a given file extension.
func (c *Config) GetGrammarForExtension(ext string) string {
	if grammar, ok := c.Grammar.Mapping[strings.ToLower(ext)]; ok {
		return grammar
	}
	return strings.TrimPrefix(ext, ".")
}
