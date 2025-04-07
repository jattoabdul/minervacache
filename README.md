# Minerva In-Memory Cache

This project implements an in-memory cache with support for multiple buckets, eviction policies, and TTL. 
The cache is limited to 255 keys total and provides gRPC & HTTP API access to manage cache operations.

## Usage:

### HTTP Server
```bash
# Install to $GOPATH/bin ( you may need to restart your shell for this to take effect)
make install

# Start HTTP server on localhost:8080
minervacache server
```

#### Endpoints
- **Health Check**: `GET /health`
- **Set**: `PUT /cache/<bucket>/<key>` (with optional query params for TTL and eviction policy)
- **Get**: `GET /cache/<bucket>/<key>`
- **Delete**: `DELETE /cache/<bucket>/<key>`
- **Statistics**: `GET /stats` (returns cache statistics using Prometheus metrics)

#### Example Usage (With curl)
```bash
# Health check
curl -X GET http://localhost:8080/health

# Set a key
curl -X PUT http://localhost:8080/cache/bucket1/key1 -d "value1"

# Set a key with TTL and eviction policy
curl -X PUT "http://localhost:8080/cache/bucket1/key1?policy=lru&ttl=1s" -d "value1"

# Get a key
curl -X GET http://localhost:8080/cache/bucket1/key1

# Delete a key
curl -X DELETE http://localhost:8080/cache/bucket1/key1
```

### Docker
You can build and run the server using Docker:
```bash
# Build the Docker image
make docker-build
# Run the Docker container to install the command line tool.
make docker-run

# Start HTTP server on localhost:8080
minervacache server
```

## Solution Approach


## TODO:
Cache:
- [x] Create the interface for the cache
- [x] Implement the in-memory cache
    - [x] Implement the cache with buckets and with a maximum of 255 keys (Entire cache is limited to 255 keys not just per bucket right?)
    - [x] Implement the eviction policies (each bucket/operation can have its own eviction policy, managed with an options attribute maybe?)
    - [x] Implement the TTL for the cache operations (cleanups done inline or run in background?)
    - [x] Write unit tests for the cache

Server:
- [ ] Implement the HTTP server
    - [x] Implement the endpoints for the cache operations (set, get, delete, stats)
    - [x] Implement the statistics endpoint
    - [x] Implement the Performance metrics collection (Prometheus?)
    - [x] Implement the health check endpoint
    - [x] Implement the graceful shutdown
    - [ ] Write unit tests for the HTTP server
- [ ] Implement the gRPC server
  - [ ] Implement the proto file for the cache operations (set, get, delete, stats)
  - [ ] Implement the gRPC server and its handlers
  - [ ] Write unit tests for the gRPC server

Client:
- [ ] Implement a client to test the gRPC server (Maybe a simple CLI client to test the servers)
- [ ] Write integration tests for both the gRPC and HTTP servers (likely using the client)

Misc:
- [ ] Write a Dockerfile for the server and setup its usage (include docker-compose)
- [x] Write a Makefile for automating the build and run process
- [ ] Write a detailed documentation in README for the project with usage instructions of the client and servers.

