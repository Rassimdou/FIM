package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/Rassimdou/FIM/agent"
	"github.com/Rassimdou/FIM/proto"
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

	//create watcher
	watcher, err := agent.NewWatcher(cfg, func(ev *proto.FileEvent) {
		log.Printf("Event: path=%s type=%s ", ev.FilePath, ev.EventType.String())
	})
	if err != nil {
		log.Fatalf("failed to create watcher: %v", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	//start watching

	if err := watcher.Start(ctx); err != nil {
		log.Fatalf("watcher error: %v", err)
	}

}
