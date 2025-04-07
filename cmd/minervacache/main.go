package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/jattoabdul/minervacache/cache"
	"github.com/jattoabdul/minervacache/server"
)

var (
	// server flags
	port int
	host string
)

func main() {
	rootCommand := &cobra.Command{
		Use:   "minervacache",
		Short: "An in-memory cache test http server",
	}

	serverCommand := &cobra.Command{
		Use:   "server",
		Short: "Start the cache server",
		Run:   runServer,
	}

	// Flags for server command
	serverCommand.Flags().IntVar(&port, "port", 8080, "Port our server listens on")
	serverCommand.Flags().StringVar(&host, "host", "0.0.0.0", "Host address our server binds to")

	rootCommand.AddCommand(serverCommand)

	if err := rootCommand.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func runServer(cmd *cobra.Command, args []string) {
	//Init prometheus metrics
	metrics := cache.NewPmMetrics()

	// Create a new cache instance
	mCache := cache.NewMinervaCache(cache.MaxCacheSize, cache.DefaultCleanupInterval, metrics)

	// Create a new server instance
	mServer := server.NewHTTPServer(mCache, metrics)
	//mServer.server

	// Setup graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start the server in a goroutine
	log.Printf("Starting minervacache HTTP server on port %s:%d\n", host, port)
	go func() {
		if err := mServer.Start(context.Background(), host, port); err != nil {
			log.Printf("Failed to start server with error: %v\n", err)
		}
	}()

	// Wait for termination signal
	sig := <-sigCh
	log.Printf("Received signal %v, shutting down gracefully\n", sig)

	// Stop the cache
	mCache.Stop()
	log.Printf("Cache stopped successfully\n")

	// Stop the server
	if err := mServer.Stop(context.Background()); err != nil {
		log.Printf("Failed to stop server with error: %v\n", err)
	} else {
		log.Printf("Server stopped successfully\n")
	}
}
