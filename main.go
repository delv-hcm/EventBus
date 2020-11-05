package main

import (
	"log"
	"net"

	"eventbus/EventBus"
	"eventbus/EventHandlers"

	"google.golang.org/grpc"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	eventBus := (&EventBus.EventBus{ClientId: "server-01"}).New()

	eventBus.QueueSubscribe(&EventHandlers.RegisterEvent{}, []EventBus.Handler{
		{
			EventHandler:    &EventHandlers.RegisterEventHandler{},
			FallbackHandler: &EventHandlers.RegisterEventHandler{},
		},
	}, "register-event-100", "queue-name",
	)

	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalln("failed to listen", "error", err)
	}
	s := grpc.NewServer()
	log.Println("GRPC server listening on", "Port", "8080")
	if err := s.Serve(lis); err != nil {
		log.Fatalln("failed to serve", "error", err)
	}
}
