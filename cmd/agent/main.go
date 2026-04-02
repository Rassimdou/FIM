package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Rassimdou/FIM/agent"
)

func main() {

	cfg, err := agent.LoadConfig("config/gowatch.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	//resolve host if ID is 'auto'
	if cfg.Agent.ID == "auto" {
		hostname, err := os.Hostname()
		if err != nil {
			log.Fatalf("failed to get hostname: %v", err)
		}
		cfg.Agent.ID = hostname
	}
	fmt.Printf("Agent ID   : %s\n", cfg.Agent.ID)
	fmt.Printf("Server     : %s\n", cfg.Server.Address)
	fmt.Printf("Watch paths: %v\n", cfg.Watch.Paths)

}
