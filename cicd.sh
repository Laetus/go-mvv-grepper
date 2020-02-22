#!/usr/bin/env bash
set -x 
BUILD_NAME=$RANDOM

docker build -t $BUILD_NAME . 
if [ -z "$(git status --porcelain)" ]; then 
  IMAGE_NAME="laetus/go-mvv-grepper:$(git rev-parse HEAD)"
else 
  IMAGE_NAME="laetus/go-mvv-grepper:temp-$(git rev-parse HEAD)"
fi

docker tag $BUILD_NAME $IMAGE_NAME
docker push $IMAGE_NAME