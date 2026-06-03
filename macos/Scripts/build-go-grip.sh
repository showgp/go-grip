#!/bin/bash
set -euo pipefail

export PATH="/opt/homebrew/bin:/usr/local/go/bin:$PATH"

# Build go-grip universal binary (arm64 + x86_64)
cd "${SRCROOT}/.."

echo "Building go-grip for arm64..."
GOOS=darwin GOARCH=arm64 go build -o /tmp/go-grip-arm64 .

echo "Building go-grip for amd64..."
GOOS=darwin GOARCH=amd64 go build -o /tmp/go-grip-amd64 .

echo "Creating universal binary..."
mkdir -p "${BUILT_PRODUCTS_DIR}/${CONTENTS_FOLDER_PATH}/Resources"
lipo -create /tmp/go-grip-arm64 /tmp/go-grip-amd64 \
     -output "${BUILT_PRODUCTS_DIR}/${CONTENTS_FOLDER_PATH}/Resources/go-grip"

chmod +x "${BUILT_PRODUCTS_DIR}/${CONTENTS_FOLDER_PATH}/Resources/go-grip"

echo "go-grip universal binary built successfully."
