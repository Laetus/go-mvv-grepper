#!/bin/bash
echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
docker push laetus/go-mvv-grepper:$TRAVIS_COMMIT
docker logout