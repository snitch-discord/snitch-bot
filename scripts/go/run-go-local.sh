#!/bin/bash

BASE_DIR=$(dirname "$0")/../..

source "${BASE_DIR}/scripts/go/token.env"

export SNITCH_BACKEND_HOST=localhost
export SNITCH_BACKEND_PORT=4200
export SNITCH_DISCORD_TOKEN="${SNITCH_DISCORD_TOKEN}"

go run "${BASE_DIR}"/cmd/snitchbot/main.go
