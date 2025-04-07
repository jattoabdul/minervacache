package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/jattoabdul/minervacache/cache"
	"github.com/jattoabdul/minervacache/proto"
	"github.com/jattoabdul/minervacache/server"
)

var (
	// server flags
	useGRPC bool

	port int
	host string

	// client flags
	gRPCPort int
	gRPCHost string
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

	grpcClientCommand := &cobra.Command{
		Use:   "client",
		Short: "Start an interactive client to test the gRPC server",
		Run:   runGRPCClient,
	}

	// Flags for server command
	serverCommand.Flags().BoolVar(&useGRPC, "grpc", false, "Use the gRPC server not the default HTTP server")
	serverCommand.Flags().IntVar(&port, "port", 8080, "Port our server listens on")
	serverCommand.Flags().StringVar(&host, "host", "0.0.0.0", "Host address our server binds to")

	// Flags for gRPC client command
	grpcClientCommand.Flags().StringVar(&gRPCHost, "host", "localhost", "Server host to connect to")
	grpcClientCommand.Flags().IntVar(&gRPCPort, "port", 8080, "Server port to connect to")

	rootCommand.AddCommand(serverCommand, grpcClientCommand)

	if err := rootCommand.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error Occured: %v\n", err)
		os.Exit(1)
	}
}

// runServer starts the cache server with the specified host and port.
// If useGRPC is true, it starts a gRPC server; otherwise, it starts an HTTP server.
func runServer(cmd *cobra.Command, args []string) {
	//Init prometheus metrics
	metrics := cache.NewPmMetrics()

	// Create a new cache instance
	mCache := cache.NewMinervaCache(cache.MaxCacheSize, cache.DefaultCleanupInterval, metrics)

	// Create a new server instance based on the useGRPC flag
	var mServer server.Server
	serverType := "HTTP"
	if useGRPC {
		serverType = "gRPC"
		mServer = server.NewGRPCServer(mCache, metrics)
	} else {
		mServer = server.NewHTTPServer(mCache, metrics)
	}
	//mServer.server

	// Setup graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Start the server in a goroutine
	log.Printf("Starting minervacache %s server on port %s:%d\n", serverType, host, port)
	go func() {
		if err := mServer.Start(context.Background(), host, port); err != nil {
			log.Printf("Failed to start server with error: %v\n", err)
		}
	}()

	// Wait for termination signal
	sig := <-sigCh
	log.Printf("Received signal %v, shutting down gracefully...\n", sig)

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

// runGRPCClient starts an interactive gRPC client to test the gRPC server.
func runGRPCClient(cmd *cobra.Command, args []string) {
	addr := fmt.Sprintf("%s:%d", gRPCHost, gRPCPort)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("gRPC Clint failed to connect to server: %v", err)
	}
	defer conn.Close()

	client := proto.NewMinervaCacheClient(conn)

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("MinervaCache CLI (type 'help' for commands, 'exit' to quit)")
	fmt.Printf("Connected to %s\n", addr)

	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		args := strings.Fields(input)
		cmd := strings.ToLower(args[0])

		switch cmd {
		case "exit", "quit":
			return
		case "help":
			printHelp()
		case "get":
			if len(args) != 3 {
				fmt.Println("Usage: get <bucket> <key>")
				continue
			}
			handleGet(client, args[1], args[2])
		case "set":
			if len(args) < 4 {
				fmt.Println("Usage: set <bucket> <key> <value> [ttl_ms]")
				continue
			}

			var ttl int64 = 0
			if len(args) > 4 {
				ttl, err = parseTTL(args[4])
				if err != nil {
					fmt.Printf("Invalid TTL value: %v\n", err)
					continue
				}
			}

			handleSet(client, args[1], args[2], args[3], ttl)
		case "del", "delete":
			if len(args) != 3 {
				fmt.Println("Usage: del <bucket> <key>")
				continue
			}

			handleDelete(client, args[1], args[2])
		default:
			fmt.Printf("Unknown command: %s\n", cmd)
			printHelp()
		}
	}
}

// parseTTL parses the TTL value from a string to an int64.
func parseTTL(ttlStr string) (int64, error) {
	ttl, err := strconv.ParseInt(ttlStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid TTL: %v", err)
	}
	return ttl, nil
}

// handleGet processes a get request
func handleGet(client proto.MinervaCacheClient, bucket, key string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &proto.GetRequest{
		Bucket: bucket,
		Key:    key,
	}
	resp, err := client.Get(ctx, req)

	if err != nil {
		fmt.Printf("Error getting value: %v\n", err)
		return
	}

	if resp.Value != nil {
		fmt.Printf("Value: %s\n", string(resp.Value))
	} else {
		fmt.Println("Error Occurred: Key not found")
	}
}

// handleSet processes a set request
func handleSet(client proto.MinervaCacheClient, bucket, key, value string, ttl int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &proto.SetRequest{
		Bucket: bucket,
		Key:    key,
		Value:  []byte(value),
		TtlMs:  int32(ttl),
	}
	resp, err := client.Set(ctx, req)

	if err != nil {
		fmt.Printf("Error setting value: %v\n", err)
		return
	}

	if resp != nil {
		fmt.Println("Value set successfully")
	} else {
		fmt.Println("Error Occurred: Value not set")
	}
}

// handleDelete processes a delete request
func handleDelete(client proto.MinervaCacheClient, bucket, key string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &proto.DeleteRequest{
		Bucket: bucket,
		Key:    key,
	}
	resp, err := client.Delete(ctx, req)

	if err != nil {
		fmt.Printf("Error deleting value: %v\n", err)
		return
	}

	if resp != nil {
		fmt.Println("Value deleted successfully")
	} else {
		fmt.Println("Error Occurred: Value not deleted")
	}
}

func printHelp() {
	fmt.Println("Available commands for Minerva gRPC client:")
	fmt.Println("  get <bucket> <key>                    Get value by bucket and key")
	fmt.Println("  set <bucket> <key> <value> [ttl_ms]  Set value with optional TTL in milliseconds")
	fmt.Println("  del <bucket> <key>                    Delete value by bucket and key")
	fmt.Println("  help                                  Show this help message")
	fmt.Println("  exit                                  Exit the client")
}
