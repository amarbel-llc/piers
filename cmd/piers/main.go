package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/amarbel-llc/purse-first/libs/go-mcp/server"
	"github.com/amarbel-llc/purse-first/libs/go-mcp/transport"
	"github.com/amarbel-llc/piers/internal/google"
	"github.com/amarbel-llc/piers/internal/tools"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	client, err := google.NewClient(ctx)
	if err != nil {
		log.Fatalf("creating google client: %v", err)
	}

	app := tools.RegisterAll(client)

	registry := server.NewToolRegistry()
	app.RegisterMCPTools(registry)

	t := transport.NewStdio(os.Stdin, os.Stdout)

	srv, err := server.New(t, server.Options{
		ServerName:    app.Name,
		ServerVersion: app.Version,
		Tools:         registry,
	})
	if err != nil {
		log.Fatalf("creating server: %v", err)
	}

	if err := srv.Run(ctx); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
