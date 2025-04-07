# Minerva In-Memory Cache

This project implements an in-memory cache with support for multiple buckets, eviction policies, and TTL. 
The cache is limited to 255 keys total and provides gRPC & HTTP API access to manage cache operations.

## Usage:

### HTTP Server
```bash
# Install to $GOPATH/bin (you may need to restart your shell for this to take effect)
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

## gRPC Server
```bash
# Install to $GOPATH/bin (if you have not already done so)
make install

# Start gRPC server on localhost:8080
minervacache server --grpc
```

#### Example Usage (With REPL)
```bash
# Start the gRPC server
minervacache server --grpc

# Start the REPL client
minervacache client

> set key1 value1
Usage: set <bucket> <key> <value> [ttl_ms]

> set bucket1 key1 value1
Value set successfully

> get bucket1 key1
Value: value1

> get bucket1 key2
Error getting value: rpc error: code = Unknown desc = key not found

> delete bucket1 key1
Value deleted successfully

> get bucket1 key1
Error getting value: rpc error: code = Unknown desc = bucket not found

> exit

```

### Docker
You can build and run the HTTP server using Docker:
```bash
# Build the Docker image
make docker-build
# Run the Docker container to install the command line tool.
make docker-run

# Start HTTP server on localhost:8080
minervacache server
```

## Solution Approach

Using a bucketed cache with a maximum of 255 keys, the cache is designed to be simple and efficient.
The cache supports multiple buckets, and allows for different eviction policies and TTLs when getting or setting keys in a bucket.
The cache is implemented as a map of buckets, where each bucket is a map of keys to values.
The cache uses a linked list to keep track of the order of keys in each bucket, allowing for efficient eviction of keys based on the specified eviction policy.
The cache also supports TTLs, allowing keys to expire after a specified time.
A weird quirk of the cache is that it supports multiple eviction policies, but we are using a single linked list to keep track of the order of keys in each bucket.
This means that the LRU and Newest policies are not strictly enforced.
Normally, the expectation is that a cache uses the same eviction policy across all buckets in the cache.
We could use two linked lists to keep track of the order of keys in each bucket, one for LRU/MRU and one for Newest/Oldest, but this would add complexity to the implementation.
The cache does a background cleanup of expired keys, to avoid scanning the entire cache during normal operations. However, the Get operation always checks for expired keys, so the cache is always up to date.

### Eviction Policies
The cache supports four eviction policies:
1. **Oldest**: Removes the item that was first added to the cache
2. **Newest**: Removes the item that was most recently added to the cache
3. **LRU** (Least Recently Used): Removes the item that hasn't been accessed for the longest time
4. **MRU** (Most Recently Used): Removes the item that was most recently accessed

### Development Notes:
- To generate or regenerate the protobuf files after creating or changing the proto, you can use the following commands:
```bash
# Install the required tools
brew install protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate the protobuf files
make proto
````


### Testing
The cache is tested using unit tests for the cache operations.
The tests cover the basic functionality of the cache, including setting, getting, and deleting keys, as well as testing some eviction policies and TTLs.

The unit and integration tests for the HTTP server is yet to be implemented.

```bash
# Run all tests
make test
```

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
- [x] Write a Dockerfile for the server and setup its usage (include docker-compose)
- [x] Write a Makefile for automating the build and run process
- [x] Write a detailed documentation in README for the project with usage instructions of the client and servers.

