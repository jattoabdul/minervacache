FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o minervacache cmd/minervacache/main.go

FROM alpine:latest AS runner

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/minervacache /app/

USER appuser

EXPOSE 8080

CMD ["/app/minervacache"]
