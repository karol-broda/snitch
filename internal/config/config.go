package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Defaults DefaultConfig `mapstructure:"defaults"`
}

// DefaultConfig contains default values for CLI options
type DefaultConfig struct {
	Interval     string   `mapstructure:"interval"`
	Numeric      bool     `mapstructure:"numeric"`
	Fields       []string `mapstructure:"fields"`
	Theme        string   `mapstructure:"theme"`
	Units        string   `mapstructure:"units"`
	Color        string   `mapstructure:"color"`
	Resolve      bool     `mapstructure:"resolve"`
	DNSCache     bool     `mapstructure:"dns_cache"`
	IPv4         bool     `mapstructure:"ipv4"`
	IPv6         bool     `mapstructure:"ipv6"`
	NoHeaders    bool     `mapstructure:"no_headers"`
	OutputFormat string   `mapstructure:"output_format"`
	SortBy       string   `mapstructure:"sort_by"`
}

var globalConfig *Config

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	if globalConfig != nil {
		return globalConfig, nil
	}

	v := viper.New()
	
	// set config name and file type (auto-detect based on extension)
	v.SetConfigName("snitch")
	// don't set config type - let viper auto-detect based on file extension
	// this allows both .toml and .yaml files to work
	v.AddConfigPath("$HOME/.config/snitch")
	v.AddConfigPath("$HOME/.snitch")
	v.AddConfigPath("/etc/snitch")

	// Environment variables
	v.SetEnvPrefix("SNITCH")
	v.AutomaticEnv()
	
	// environment variable bindings for readme-documented variables
	_ = v.BindEnv("config", "SNITCH_CONFIG")
	_ = v.BindEnv("defaults.resolve", "SNITCH_RESOLVE")
	_ = v.BindEnv("defaults.dns_cache", "SNITCH_DNS_CACHE")
	_ = v.BindEnv("defaults.theme", "SNITCH_THEME")
	_ = v.BindEnv("defaults.color", "SNITCH_NO_COLOR")
	
	// Set defaults
	setDefaults(v)
	
	// Handle SNITCH_CONFIG environment variable for custom config path
	if configPath := os.Getenv("SNITCH_CONFIG"); configPath != "" {
		v.SetConfigFile(configPath)
	}
	
	// Try to read config file
	if err := v.ReadInConfig(); err != nil {
		// It's OK if config file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}
	
	// Handle special environment variables
	handleSpecialEnvVars(v)
	
	// Unmarshal into config struct
	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}
	
	globalConfig = config
	return config, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("defaults.interval", "1s")
	v.SetDefault("defaults.numeric", false)
	v.SetDefault("defaults.fields", []string{"pid", "process", "user", "proto", "state", "laddr", "lport", "raddr", "rport"})
	v.SetDefault("defaults.theme", "auto")
	v.SetDefault("defaults.units", "auto")
	v.SetDefault("defaults.color", "auto")
	v.SetDefault("defaults.resolve", true)
	v.SetDefault("defaults.dns_cache", true)
	v.SetDefault("defaults.ipv4", false)
	v.SetDefault("defaults.ipv6", false)
	v.SetDefault("defaults.no_headers", false)
	v.SetDefault("defaults.output_format", "table")
	v.SetDefault("defaults.sort_by", "")
}

func handleSpecialEnvVars(v *viper.Viper) {
	// Handle SNITCH_NO_COLOR - if set to "1", disable color
	if os.Getenv("SNITCH_NO_COLOR") == "1" {
		v.Set("defaults.color", "never")
	}
	
	// Handle SNITCH_RESOLVE - if set to "0", disable resolution
	if os.Getenv("SNITCH_RESOLVE") == "0" {
		v.Set("defaults.resolve", false)
		v.Set("defaults.numeric", true)
	}

	// Handle SNITCH_DNS_CACHE - if set to "0", disable dns caching
	if os.Getenv("SNITCH_DNS_CACHE") == "0" {
		v.Set("defaults.dns_cache", false)
	}
}

// Get returns the global configuration, loading it if necessary
func Get() *Config {
	if globalConfig == nil {
		config, err := Load()
		if err != nil {
			return &Config{
				Defaults: DefaultConfig{
					Interval:     "1s",
					Numeric:      false,
					Fields:       []string{"pid", "process", "user", "proto", "state", "laddr", "lport", "raddr", "rport"},
					Theme:        "auto",
					Units:        "auto",
					Color:        "auto",
					Resolve:      true,
					DNSCache:     true,
					IPv4:         false,
					IPv6:         false,
					NoHeaders:    false,
					OutputFormat: "table",
					SortBy:       "",
				},
			}
		}
		return config
	}
	return globalConfig
}

// GetInterval returns the configured interval as a duration
func (c *Config) GetInterval() time.Duration {
	if duration, err := time.ParseDuration(c.Defaults.Interval); err == nil {
		return duration
	}
	return time.Second // default fallback
}

// CreateExampleConfig creates an example configuration file
func CreateExampleConfig(path string) error {
	exampleConfig := `# snitch configuration file
# See https://github.com/you/snitch for full documentation

[defaults]
# Default refresh interval for watch/stats/trace commands
interval = "1s"

# Disable name/service resolution by default
numeric = false

# Default fields to display (comma-separated list)
fields = ["pid", "process", "user", "proto", "state", "laddr", "lport", "raddr", "rport"]

# Default theme for TUI (dark, light, mono, auto)
theme = "auto"

# Default units for byte display (auto, si, iec)
units = "auto"

# Default color mode (auto, always, never)
color = "auto"

# Enable name resolution by default
resolve = true

# Filter options
ipv4 = false
ipv6 = false

# Output options
no_headers = false
output_format = "table"
sort_by = ""
`

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// Write config file
	if err := os.WriteFile(path, []byte(exampleConfig), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}