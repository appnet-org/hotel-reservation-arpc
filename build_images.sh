#!/bin/bash

# Go to the directory containing this script (where Dockerfile is)
cd "$(dirname "$0")"

EXEC="docker buildx"
USER="appnetorg"
TAG="latest"

# Create builder only if it does not already exist
if ! $EXEC inspect mybuilder >/dev/null 2>&1; then
  $EXEC create --name mybuilder --use
else
  $EXEC use mybuilder
fi

IMAGE="hotel-reservation-arpc"
echo "Processing image ${IMAGE}"
$EXEC build -t "$USER"/"$IMAGE":"$TAG" -f Dockerfile . --push
echo
