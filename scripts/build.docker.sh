#!/bin/bash
set -eu

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"
GIT_COMMIT="${GIT_COMMIT:-$(cd ${SCRIPT_DIR} && git rev-parse --short HEAD)}"
IMAGE_NAME="redis-elasticsearch-go-example/app"

cd ${SCRIPT_DIR}/..
go mod vendor
cd -

docker build \
  -f ${SCRIPT_DIR}/Dockerfile \
  --build-arg GIT_COMMIT=${GIT_COMMIT} \
  -t $IMAGE_NAME \
  ${SCRIPT_DIR}/..
