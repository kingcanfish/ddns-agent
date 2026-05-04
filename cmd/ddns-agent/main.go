package main

import (
	"log"
	"os"

	"github.com/kingcanfish/ddns-agent/internal/agent"
	"github.com/kingcanfish/ddns-agent/internal/config"
)

func main() {
	configPath := "config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Config loaded: domain=%s", cfg.Domain)

	a, err := agent.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	a.Run()
}
