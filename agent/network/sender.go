package network

import (
	"context"
	"log"
	"time"

	"github.com/Rassimdou/FIM/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Sender manages the gRPC connection and an internal queue of events.
type Sender struct {
	serverAddr string
	eventQueue chan *proto.FileEvent
}

// NewSender creates a new Sender with a buffered queue.
func NewSender(serverAddr string, queueSize int) *Sender {
	return &Sender{
		serverAddr: serverAddr,
		// Create a buffered channel to hold events
		eventQueue: make(chan *proto.FileEvent, queueSize),
	}
}

// Enqueue attempts to add an event to the queue. If full, it drops the event to prevent blocking the file watcher.
func (s *Sender) Enqueue(ev *proto.FileEvent) {
	select {
	case s.eventQueue <- ev:
		//Success ,The event slid right into the channel.
	default:
		//the channel is completely full. We must drop the event.
		log.Printf("WARNING: Network queue full! Dropping event for %s", ev.FilePath)
	}
}

/*
start runs a continuous loop that attempts to connect to the server
and send events , if the connection drops, it automatically
awaits and tried indefinitely until the server is back online
This ensures that the agent can recover from temporary network issues without losing all events
*/
func (s *Sender) Start(ctx context.Context) {

	for {
		err := s.connectAndSend(ctx)
		if ctx.Err() != nil {
			log.Println("Sender stopping gracefully...")
			return
		}
		if err != nil {
			log.Printf("Network connection lost: %v. Retrying in 5 seconds...", err)
		}
		time.Sleep(5 * time.Second)
	}
}

func (s *Sender) connectAndSend(ctx context.Context) error {
	log.Printf("Attempting to connect to server at %s", s.serverAddr)

	//Dial the server (this is non-blocking in gRPC by default)
	conn, err := grpc.NewClient(s.serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	client := proto.NewFileEventServiceClient(conn)

	//open bi-directional stream
	stream, err := client.SendFileEvent(ctx)
	if err != nil {
		return err
	}

	log.Println("Connected to server. Starting to send events...")

	//now server back online we loop through the event queue and send events to the server

	for {
		select {
		//user press ctrl+c or kill the process
		case <-ctx.Done():
			stream.CloseAndRecv()
			return nil

			//new event to send
		case ev := <-s.eventQueue:

			//we try to send the event to the server
			if err := stream.Send(ev); err != nil {
				log.Printf("Error sending event to server: %v", err)

				//CRITICAL: put the event back into the queue so we dont lose it
				s.Enqueue(ev)

				return err
			}

		}
	}
}
