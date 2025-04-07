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
