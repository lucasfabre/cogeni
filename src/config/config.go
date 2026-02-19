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

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
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
	URL string `mapstructure:"url" json:"url"`
	// Branch is the specific Git branch to use (e.g., "master", "main", "develop").
	Branch string `mapstructure:"branch" json:"branch"`
	// BuildCmd is an optional shell command to execute for building the grammar.
	// If omitted, the manager will attempt a standard C/C++ compilation.
	BuildCmd string `mapstructure:"build_cmd" json:"build_cmd"`
	// Artifact is the name or relative path of the resulting shared library
	// (e.g., "python.so", "build/sql.dylib").
	Artifact string `mapstructure:"artifact" json:"artifact"`
}

// Config is the root configuration object for the cogeni application.
type Config struct {
	// Grammar contains all settings related to Tree-sitter grammars.
	Grammar struct {
		// Location is the filesystem path where grammar shared libraries are cached.
		Location string `mapstructure:"location" json:"location"`
		// Mapping maps file extensions (e.g., ".py") to grammar names (e.g., "python").
		Mapping map[string]string `mapstructure:"mapping" json:"mapping"`
		// Sources provides custom build instructions for specific grammar names.
		Sources map[string]GrammarSource `mapstructure:"sources" json:"sources"`
	} `mapstructure:"grammar" json:"grammar"`
}

// LoadConfig initializes the application configuration.
// It searches for "config.yaml" in standard OS-specific configuration directories
// (e.g., ~/.config/cogeni/ on Linux) and applies defaults and environment variable overrides.
// Environment variables follow the pattern COGENI_PATH_TO_KEY (e.g., COGENI_GRAMMAR_LOCATION).
func LoadConfig() (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName("config")

	// Determine config file path based on OS and XDG standards
	configPath, err := getConfigPath()
	if err != nil {
		return nil, fmt.Errorf("failed to determine config path: %w", err)
	}
	v.AddConfigPath(configPath)

	// Set defaults
	defaultGrammarLocation, err := getDefaultGrammarLocation()
	if err != nil {
		return nil, fmt.Errorf("failed to determine default grammar location: %w", err)
	}
	v.SetDefault("grammar.location", defaultGrammarLocation)
	v.SetDefault("grammar.mapping", map[string]string{
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
	})

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, defaults will be used.
			fmt.Fprintf(os.Stderr, "Config file not found in %s, using defaults.\n", configPath)
		} else {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Environment variable overrides
	v.SetEnvPrefix(strings.ToUpper(appName)) // e.g., COGENI_GRAMMAR_LOCATION
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// getConfigPath determines the configuration directory based on OS.
func getConfigPath() (string, error) {
	var configDir string
	home, err := homedir.Dir()
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
	home, err := homedir.Dir()
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
