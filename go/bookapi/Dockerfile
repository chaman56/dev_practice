# Stage 1: build the binary using the full Go toolchain.
FROM golang:1.24 AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
# CGO_ENABLED=0 makes a fully static binary that runs anywhere.
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /bin/api ./cmd/api

# Stage 2: copy only the binary into a tiny, secure image (no shell, no compiler).
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /bin/api /api
EXPOSE 8080
USER nonroot
ENTRYPOINT ["/api"]
