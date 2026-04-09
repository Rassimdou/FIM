package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/Rassimdou/FIM/agent"
	"github.com/Rassimdou/FIM/agent/network"
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

	//create context to handle graceful shutdown on Ctrl+C
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	//initialize our resilient network sender(queue size: 1000)
	sender, err := network.NewSender(
		cfg.Server.Address,
		1000,
		cfg.Server.CACert,
		cfg.Server.ClientCert,
		cfg.Server.ClientKey,
	)
	if err != nil {
		log.Fatalf("failed to initialize network sender (TLS error): %v", err)
	}

	go sender.Start(ctx)

	//create watcher and tell it to send events to the queue
	watcher, err := agent.NewWatcher(cfg, func(ev *proto.FileEvent) {

		log.Printf("Local Event: path=%s type=%s", ev.FilePath, ev.EventType.String())

		//safely put the event in the queue (never blocks)
		sender.Enqueue(ev)

	})
	if err != nil {
		log.Fatalf("failed to create watcher: %v", err)
	}

	//start the baseline scan asynchronously so it doesn't block the watcher
	go func() {
		log.Println("Initiating baseline scan...")
		if err := agent.ScanExistingFiles(ctx, cfg, func(ev *proto.FileEvent) {
			sender.Enqueue(ev)
		}); err != nil {
			log.Printf("baseline scan failed: %v", err)
		}
	}()

	//start watching
	log.Println("Agent started. Watching for file changes...")
	if err := watcher.Start(ctx); err != nil {
		log.Fatalf("watcher error: %v", err)
	}
}
