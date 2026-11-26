# Build stage: compile the Go backend
FROM golang:1.23-alpine AS builder
WORKDIR /app
RUN apk add --no-cache gcc musl-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /todo-server .

# Run stage: minimal image
FROM alpine:3.21
WORKDIR /app
EXPOSE 8080
COPY --from=builder /todo-server /todo-server
# Wait-for-it is not included; the entrypoint relies on compose health/depends_on.
CMD ["/todo-server"]
