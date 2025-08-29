# Use the official Golang image as the base image
FROM golang:1.22.1-bullseye AS builder
ENV CGO_ENABLED=1

# Set the working directory
WORKDIR /workspace

# Copy the rest of the source code
COPY cmd/ cmd/
COPY proto/ proto/
COPY config.json config.json
COPY dialer/ dialer/
COPY registry/ registry/
COPY services/ services/
COPY tls/ tls/
COPY tracing/ tracing/
COPY tune/ tune/
COPY go-lib/interceptor go-lib/interceptor

# Copy the go.mod and go.sum files to the workspace
COPY go.mod go.sum ./

# Download dependencies (caching these steps speeds up subsequent builds)
RUN go mod download

# Build the Go binaries
RUN go install -ldflags="-s -w" -trimpath ./cmd/...


# FROM gcr.io/distroless/static:nonroot
# FROM ubuntu:20.04
FROM alpine:latest
RUN apk add gcompat

WORKDIR /

COPY --from=builder /workspace/config.json .
COPY --from=builder /go/bin/frontend .
COPY --from=builder /go/bin/geo .
COPY --from=builder /go/bin/profile .
COPY --from=builder /go/bin/rate .
COPY --from=builder /go/bin/recommendation .
COPY --from=builder /go/bin/reservation .
COPY --from=builder /go/bin/search .
COPY --from=builder /go/bin/user .
