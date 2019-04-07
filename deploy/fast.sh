#!/bin/bash -eu

go test ./...

CGO_ENABLED=0 GOOS=linux go build -o diagnose-binary -v

TAG_DEST=$(git rev-parse --short HEAD)

docker build -f deploy/Dockerfile -t gcr.io/eoscanada-shared-services/diagnose:$TAG_DEST .

docker tag  gcr.io/eoscanada-shared-services/diagnose:$TAG_DEST gcr.io/eoscanada-shared-services/diagnose:latest

echo "--- Pushing tag $TAG_DEST"
gcloud docker -- push gcr.io/eoscanada-shared-services/diagnose:$TAG_DEST
echo "--- Pushing tag latest"
gcloud docker -- push gcr.io/eoscanada-shared-services/diagnose:latest
