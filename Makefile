.PHONY: build test run macos macos-run macos-test macos-archive macos-dmg macos-candidate macos-candidate-check macos-clean

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

CANDIDATE_ARCHIVE := macos/.build/GoGrip.xcarchive
CANDIDATE_DIR := macos/.build/candidate
CANDIDATE_APP := $(CANDIDATE_ARCHIVE)/Products/Applications/GoGrip.app

.PHONY: macos-archive
macos-archive:  ## Archive the candidate app (Release, universal host + bundled Go, ad-hoc signed)
	rm -rf "$(CANDIDATE_ARCHIVE)"
	xcodebuild -project macos/GoGrip.xcodeproj -scheme GoGrip \
	  -configuration Release -archivePath "$(CANDIDATE_ARCHIVE)" archive

.PHONY: macos-dmg
macos-dmg: macos-archive  ## Build the candidate DMG from the archived app (with /Applications entry)
	rm -rf "$(CANDIDATE_DIR)/dmg-root"
	mkdir -p "$(CANDIDATE_DIR)/dmg-root"
	ditto "$(CANDIDATE_APP)" "$(CANDIDATE_DIR)/dmg-root/GoGrip.app"
	ln -sfn /Applications "$(CANDIDATE_DIR)/dmg-root/Applications"
	hdiutil create -volname "GoGrip" -srcfolder "$(CANDIDATE_DIR)/dmg-root" \
	  -ov -format UDZO "$(CANDIDATE_DIR)/GoGrip.dmg"

.PHONY: macos-candidate
macos-candidate: macos-dmg  ## Build the candidate App zip + DMG, then verify signatures and architectures
	ditto -c -k --sequesterRsrc --keepParent "$(CANDIDATE_APP)" "$(CANDIDATE_DIR)/GoGrip.app.zip"
	"$(MAKE)" macos-candidate-check

.PHONY: macos-candidate-check
macos-candidate-check:  ## Verify the archived candidate: ad-hoc signatures, identifiers, architectures
	@test -d "$(CANDIDATE_APP)" || { echo "missing archived app: $(CANDIDATE_APP)"; exit 1; }
	@codesign --verify --strict "$(CANDIDATE_APP)"
	@codesign --verify --strict "$(CANDIDATE_APP)/Contents/MacOS/go-grip"
	@codesign -dv "$(CANDIDATE_APP)" 2>&1 | grep -qx "Identifier=com.showgp.GoGrip"
	@codesign -dv "$(CANDIDATE_APP)/Contents/MacOS/go-grip" 2>&1 | grep -qx "Identifier=com.showgp.GoGrip.go-grip"
	@codesign -dv "$(CANDIDATE_APP)" 2>&1 | grep -q "Signature=adhoc"
	@codesign -dv "$(CANDIDATE_APP)/Contents/MacOS/go-grip" 2>&1 | grep -q "Signature=adhoc"
	@for binary in "$(CANDIDATE_APP)/Contents/MacOS/GoGrip" "$(CANDIDATE_APP)/Contents/MacOS/go-grip"; do \
	  lipo "$$binary" -verify_arch arm64 || { echo "missing arm64: $$binary"; exit 1; }; \
	  lipo "$$binary" -verify_arch x86_64 || { echo "missing x86_64: $$binary"; exit 1; }; \
	done
	@echo "candidate checks passed: $(CANDIDATE_APP)"

.PHONY: macos-clean
macos-clean:  ## Clean macOS build artifacts
	rm -rf macos/.build
