.PHONY: build test run macos macos-run macos-clean dmg

build:
	go build -o go-grip .

test:
	go test ./...

run:
	go run . $(ARGS)

.PHONY: macos
macos:  ## Build macOS App (GoGrip.app)
	xcodebuild -project macos/GoGrip.xcodeproj -scheme GoGrip \
	  -configuration Release \
	  CODE_SIGN_IDENTITY="" CODE_SIGNING_ALLOWED=NO build

.PHONY: macos-run
macos-run:  ## Build and run macOS App in Debug mode
	xcodebuild -project macos/GoGrip.xcodeproj -scheme GoGrip \
	  -configuration Debug \
	  CODE_SIGN_IDENTITY="" CODE_SIGNING_ALLOWED=NO build
	open macos/build/Debug/GoGrip.app

.PHONY: macos-test
macos-test:  ## Run macOS host adapter behavior tests (real Go, logic test bundle)
	mkdir -p macos/.build/test
	go build -o macos/.build/test/go-grip .
	rm -rf /tmp/gogrip-managed-test-tool
	mkdir -p /tmp/gogrip-managed-test-tool/GoGrip.app/Contents/MacOS
	cp macos/.build/test/go-grip /tmp/gogrip-managed-test-tool/GoGrip.app/Contents/MacOS/go-grip
	xcodebuild -project macos/GoGrip.xcodeproj -scheme GoGrip \
	  -configuration Debug -destination 'platform=macOS' \
	  CODE_SIGN_IDENTITY="" CODE_SIGNING_ALLOWED=NO test

.PHONY: macos-clean
macos-clean:  ## Clean macOS build artifacts
	xcodebuild -project macos/GoGrip.xcodeproj -scheme GoGrip \
	  -configuration Release clean
	xcodebuild -project macos/GoGrip.xcodeproj -scheme GoGrip \
	  -configuration Debug clean

.PHONY: dmg
dmg: macos  ## Create DMG from built app (requires release build)
	mkdir -p dmg-root
	cp -R macos/build/Release/GoGrip.app dmg-root/
	ln -s /Applications dmg-root/Applications
	hdiutil create -volname "GoGrip" -srcfolder dmg-root \
	  -ov -format UDZO GoGrip.dmg
	rm -rf dmg-root
