#!/bin/bash
set -euo pipefail

export PATH="/opt/homebrew/bin:/usr/local/go/bin:$PATH"

# Build the bundled go-grip tool into the app's Contents/MacOS and ad-hoc sign
# it before Xcode signs the bundle. The host loads it only from this location;
# there is no Resources, PATH or source-tree fallback at runtime.
cd "${SRCROOT}/.."

echo "Building go-grip for arm64..."
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o /tmp/go-grip-arm64 .

echo "Building go-grip for amd64..."
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o /tmp/go-grip-amd64 .

TOOL="${BUILT_PRODUCTS_DIR}/${CONTENTS_FOLDER_PATH}/MacOS/go-grip"
echo "Creating universal binary at ${TOOL}..."
mkdir -p "$(dirname "${TOOL}")"
lipo -create /tmp/go-grip-arm64 /tmp/go-grip-amd64 -output "${TOOL}"

echo "Ad-hoc signing the bundled tool..."
codesign --force --sign - --identifier com.showgp.GoGrip.go-grip "${TOOL}"

echo "go-grip universal binary built and signed successfully."
