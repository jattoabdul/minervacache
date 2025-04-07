// Package server implements HTTP server for accessing the cache over the network.
package server

import "context"

// Server is an interface for the cache server.
// Defines methods to start and stop the server and can be implemented by different server protocols e.g. HTTP, gRPC, etc.
type Server interface {
	Start(ctx context.Context, addr string, port int) error
	Stop(ctx context.Context) error
}
