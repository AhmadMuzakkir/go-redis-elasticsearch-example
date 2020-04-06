#!/bin/bash
set -eu

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"
ES_CONTAINER_NAME="redis-elasticsearch-go-example-integration-test"

cd "${SCRIPT_DIR}"/..

# Remove the container if it already exists.
EXISTING="$(docker ps -q --filter name=${ES_CONTAINER_NAME})"
if [[ ${EXISTING} ]]; then
  docker rm -vf "${EXISTING}" >/dev/null >>/dev/null
fi

docker run --name ${ES_CONTAINER_NAME} -d -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" -e "network.host=_local_,_site_" -e "network.publish_host=_local_" docker.elastic.co/elasticsearch/elasticsearch:7.6.0 >/dev/null

go test -v -covermode=atomic -tags=integration, -timeout=15m ./...

docker rm -vf "$(docker ps -q --filter name=${ES_CONTAINER_NAME})" >/dev/null
