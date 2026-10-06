#!/usr/bin/env bash
set -euo pipefail

# Verify the integration passes executable roots and production build flags.
[[ "$*" == *"--packages ./cmd/server,./cmd/helper"* ]]
[[ "$*" == *"--build-flag=-mod=vendor"* ]]
[[ "$*" == *"--build-flag=-tags=production"* ]]
[[ "$*" == *"--goos linux --goarch amd64 --cgo-enabled 0"* ]]
printf '%s\n' "$*"
printf '%s\n' '{"schemaVersion":"crypto-inventory/v1","artifacts":[]}' > ./report/source.json
