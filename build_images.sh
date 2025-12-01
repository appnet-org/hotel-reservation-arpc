#!/bin/bash

set -e

# --- Configuration ---
USER="appnetorg"
TAG="latest"
IMAGE="hotel-reservation-arpc-tcp"
UPDATE_ARPC="1"  # Set to "1" to update aRPC dependency to latest main, "0" to use pinned version
# ---

# Optionally refresh the aRPC-TCP dependency before building
if [ "$UPDATE_ARPC" = "1" ]; then
  echo "Updating aRPC-TCP dependency to latest arpc-tcp-with-element branch..."
  go get -u github.com/appnet-org/arpc-tcp@arpc-tcp-with-element
  go mod tidy
else
  echo "Using pinned aRPC-TCP version from go.mod"
fi

# Build the Docker image
echo "Building Docker image: $USER/$IMAGE:$TAG"
sudo docker build -t "$USER/$IMAGE:$TAG" -f Dockerfile .

# Push the Docker image
echo "Pushing Docker image: $USER/$IMAGE:$TAG"
sudo docker push "$USER/$IMAGE:$TAG"

echo "✅ Process complete."
