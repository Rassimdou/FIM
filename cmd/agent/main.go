package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/Rassimdou/FIM/agent"
	"github.com/Rassimdou/FIM/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	// Connect to gRPC Server
	conn, err := grpc.NewClient(cfg.Server.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	client := proto.NewFileEventServiceClient(conn)

	// Create context for the stream
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	stream, err := client.SendFileEvent(ctx)
	if err != nil {
		log.Fatalf("Failed to open stream to server: %v", err)
	}

	//create watcher
	watcher, err := agent.NewWatcher(cfg, func(ev *proto.FileEvent) {

		log.Printf("Local Event: path=%s type=%s", ev.FilePath, ev.EventType.String())

		// Send over gRPC
		if err := stream.Send(ev); err != nil {
			log.Printf("Failed to send event to server: %v", err)
		}
	})
	if err != nil {
		log.Fatalf("failed to create watcher: %v", err)
	}

	//start watching
	if err := watcher.Start(ctx); err != nil {
		log.Fatalf("watcher error: %v", err)
	}

	// Close the stream cleanly on exit
	_, err = stream.CloseAndRecv()
	if err != nil {
		log.Printf("Error closing stream: %v", err)
	}
}
