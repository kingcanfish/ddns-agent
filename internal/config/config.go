package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Dnspod        string   `yaml:"dnspod_token"`
	Domain        string   `yaml:"domain"`
	IPv4Subdomain string   `yaml:"ipv4_subdomain"`
	IPv6Subdomain string   `yaml:"ipv6_subdomain"`
	LANSubdomain  string   `yaml:"lan_subdomain"`
	IPv4URL       string   `yaml:"ipv4_url"`
	IPv6URL       string   `yaml:"ipv6_url"`
	Interval      int      `yaml:"interval"`
	Telegram      Telegram `yaml:"telegram"`
	LogLevel      string   `yaml:"log_level"`
}

type Telegram struct {
	BotToken string `yaml:"bot_token"`
	ChatID   string `yaml:"chat_id"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	cfg.setDefaults()
	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Dnspod == "" {
		return fmt.Errorf("dnspod_token is required")
	}
	if c.Domain == "" {
		return fmt.Errorf("domain is required")
	}
	if c.IPv4Subdomain == "" && c.IPv6Subdomain == "" && c.LANSubdomain == "" {
		return fmt.Errorf("at least one subdomain (ipv4_subdomain, ipv6_subdomain, lan_subdomain) is required")
	}
	return nil
}

func (c *Config) setDefaults() {
	if c.Interval <= 0 {
		c.Interval = 600
	}
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}
}
