#!/bin/bash

BASE_DIR=$(dirname "$0")/..

docker network inspect snitch-network >/dev/null 2>&1 || \
  docker network create snitch-network

bash "${BASE_DIR}"/scripts/go/run-go.sh
