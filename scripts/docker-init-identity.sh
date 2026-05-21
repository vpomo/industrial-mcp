#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
IDENTITY_DIR="$ROOT/docker/identity"
MACHINE_ID_FILE="$IDENTITY_DIR/machine-id"

mkdir -p "$IDENTITY_DIR"

if [ ! -s "$MACHINE_ID_FILE" ]; then
	uuidgen | tr -d '-' | tr '[:upper:]' '[:lower:]' >"$MACHINE_ID_FILE"
	echo "Created stable machine-id: $MACHINE_ID_FILE"
else
	echo "Using existing machine-id: $MACHINE_ID_FILE"
fi
