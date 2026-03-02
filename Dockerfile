FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod and source code
COPY go.mod go.sum ./
COPY api/ ./api/
COPY config/ ./config/
COPY pkg/ ./pkg/
COPY frontend/ ./frontend/
COPY history/ ./history/
COPY matching/ ./matching/

# We'll need gcc for sqlite if any tests use it, but alpine golang has enough for go build.
RUN go mod download

# Build all three services
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/frontend ./frontend/cmd/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/history ./history/cmd/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/matching ./matching/cmd/main.go

# Copy default config template if we want (we'll mount it via compose though)
COPY config.yaml /app/config.yaml

# Create minimal runtime image
FROM alpine:latest

WORKDIR /app
COPY --from=builder /bin/frontend /bin/frontend
COPY --from=builder /bin/history /bin/history
COPY --from=builder /bin/matching /bin/matching
# Require config is mounted or exists
COPY --from=builder /app/config.yaml /app/config.yaml

EXPOSE 8080 8081 8082

# We decide which binary to run via Docker command
CMD ["/bin/frontend"]
