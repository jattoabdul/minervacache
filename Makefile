.PHONY: proto

install: fmt
	go install ./cmd/minervacache

fmt:
	go fmt ./...

test:
	go test ./...

docker-build:
	docker build -t minervacache:latest .

docker-run:
	docker run -p 8080:8080 minervacache:latest

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/minervacache.proto
