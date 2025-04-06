# Minerva In-Memory Cache

This project implements an in-memory cache with support for multiple buckets, eviction policies, and TTL. 
The cache is limited to 255 keys total and provides gRPC & HTTP API access to manage cache operations.

## Usage:

## Solution Approach


## TODO:
Cache:
[ ] Create the interface for the cache
[ ] Implement the in-memory cache
    [ ] Implement the cache with buckets and with a maximum of 255 keys (Entire cache is limited to 255 keys not just per bucket right?)
    [ ] Implement the eviction policies (each bucket/operation can have its own eviction policy, managed with an options attribute maybe?)
    [ ] Implement the TTL for the cache operations (cleanups done inline or run in background?)
    [ ] Write unit tests for the cache

Server:
[ ] Implement the gRPC server
    [ ] Implement the proto file for the cache operations (set, get, delete, stats)
    [ ] Implement the gRPC server and its handlers
    [ ] Write unit tests for the gRPC server
[ ] Implement the HTTP server
    [ ] Implement the endpoints for the cache operations (set, get, delete, stats)
    [ ] Implement the statistics endpoint
    [ ] Implement the Performance metrics collection (Prometheus?)
    [ ] Implement the health check endpoint
    [ ] Implement the graceful shutdown
[ ] Write unit tests for the server

Client:
[ ] Implement a client to test the gRPC server (Maybe a simple CLI client to test the servers)
[ ] Write integration tests for both the gRPC and HTTP servers (likely using the client)

[ ] Write a Dockerfile for the server and setup its usage (include docker-compose)
[ ] Write a Makefile for automating the build and run process
[ ] Write a detailed documentation in README for the project with usage instructions of the client and servers.

