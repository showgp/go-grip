.PHONY: build test run macos macos-run macos-clean

build:
	go build -o go-grip .

test:
	go test ./...

run:
	go run . $(ARGS)

.PHONY: macos
macos:  ## Build macOS App (GoGrip.app, ad-hoc signed development build)
	xcodebuild -project macos/GoGrip.xcodeproj -scheme GoGrip \
	  -configuration Release -derivedDataPath macos/.build/DerivedData build

.PHONY: macos-run
macos-run:  ## Build and run macOS App in Debug mode
	xcodebuild -project macos/GoGrip.xcodeproj -scheme GoGrip \
	  -configuration Debug -derivedDataPath macos/.build/DerivedData build
	open macos/.build/DerivedData/Build/Products/Debug/GoGrip.app

.PHONY: macos-test
macos-test:  ## Run macOS host adapter behavior tests (real Go, logic test bundle)
	mkdir -p macos/.build/test
	go build -o macos/.build/test/go-grip .
	rm -rf /tmp/gogrip-managed-test-tool
	mkdir -p /tmp/gogrip-managed-test-tool/GoGrip.app/Contents/MacOS
	cp macos/.build/test/go-grip /tmp/gogrip-managed-test-tool/GoGrip.app/Contents/MacOS/go-grip
	xcodebuild -project macos/GoGrip.xcodeproj -scheme GoGrip \
	  -configuration Debug -destination 'platform=macOS' \
	  -derivedDataPath macos/.build/DerivedData test

.PHONY: macos-clean
macos-clean:  ## Clean macOS build artifacts
	rm -rf macos/.build/DerivedData
