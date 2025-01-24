#!/bin/bash

BASE_DIR=$(dirname "$0")/../..
GO_IMAGE_NAME=snitchbot

source "${BASE_DIR}/scripts/go/token.env"

docker stop ${GO_IMAGE_NAME}
docker container rm ${GO_IMAGE_NAME}

docker build -t ${GO_IMAGE_NAME} -f "${BASE_DIR}"/Containerfile .

docker run -d --name ${GO_IMAGE_NAME} \
  -e SNITCH_DISCORD_TOKEN="${SNITCH_DISCORD_TOKEN}" \
  -e SNITCH_BACKEND_HOST=snitchbe \
  -e SNITCH_BACKEND_PORT=4200 \
  ${GO_IMAGE_NAME}
