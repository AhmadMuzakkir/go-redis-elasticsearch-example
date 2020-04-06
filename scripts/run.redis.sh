#!/bin/bash
set -eu

docker run --rm -it redis:5.0.7 bash -c "docker-entrypoint.sh redis-server & bash"
