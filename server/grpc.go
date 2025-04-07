package server

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"

	"github.com/jattoabdul/minervacache/cache"
	"github.com/jattoabdul/minervacache/proto"
)

type grpcServer struct {
	proto.UnimplementedMinervaCacheServer

	cache   cache.Cache
	metrics cache.MetricsExporter
	server  *grpc.Server
}

// NewGRPCServer creates a new gRPC server with the given cache and metrics exporter.
// The server will be initialized in the Start method.
func NewGRPCServer(cache cache.Cache, metrics cache.MetricsExporter) Server {
	return &grpcServer{
		cache:   cache,
		metrics: metrics,
	}
}

// Start starts the gRPC server on the given address and port.
func (s *grpcServer) Start(ctx context.Context, addr string, port int) error {
	addr = fmt.Sprintf("%s:%d", addr, port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s.server = grpc.NewServer()
	proto.RegisterMinervaCacheServer(s.server, s)

	return s.server.Serve(listener)
}

// Stop stops the gRPC server.
func (s *grpcServer) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}
	s.server.GracefulStop()
	return nil
}

// Get handles the gRPC Get request.
func (s *grpcServer) Get(ctx context.Context, req *proto.GetRequest) (*proto.GetResponse, error) {
	mcb, err := s.cache.Get(req.Bucket, req.Key, cache.Options{})
	if err != nil {
		return nil, err
	}

	return &proto.GetResponse{Value: mcb}, nil
}

// Set handles the gRPC Set request.
func (s *grpcServer) Set(ctx context.Context, req *proto.SetRequest) (*proto.SetResponse, error) {
	// Set the value in the cache
	err := s.cache.Set(req.Bucket, req.Key, req.Value, cache.Options{})
	if err != nil {
		return nil, err
	}

	// Return an empty response
	return &proto.SetResponse{}, nil
}

// Delete handles the gRPC Delete request.
func (s *grpcServer) Delete(ctx context.Context, req *proto.DeleteRequest) (*proto.DeleteResponse, error) {
	err := s.cache.Delete(req.Bucket, req.Key)
	if err != nil {
		return nil, err
	}

	return &proto.DeleteResponse{}, nil
}
