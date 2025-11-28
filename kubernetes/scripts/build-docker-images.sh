#!/bin/bash

cd $(dirname $0)/..

EXEC="docker buildx"
USER="appnetorg"
TAG="latest"

# ENTER THE ROOT FOLDER
cd ../
ROOT_FOLDER=$(pwd)

# Create builder only if it does not already exist
if ! $EXEC inspect mybuilder >/dev/null 2>&1; then
  $EXEC create --name mybuilder --use
else
  $EXEC use mybuilder
fi

for i in hotel-reservation-arpc
do
  IMAGE=${i}
  echo "Processing image ${IMAGE}"
  cd "$ROOT_FOLDER"
  $EXEC build -t "$USER"/"$IMAGE":"$TAG" -f Dockerfile . --push
  echo
done

cd - >/dev/null
