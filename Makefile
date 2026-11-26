DIST ?= dist

.PHONY: release
release:
	@mkdir -p "$(DIST)"
	@set -e; for target in darwin/arm64 darwin/amd64 linux/arm64 linux/amd64 windows/amd64; do \
		os=$${target%/*}; arch=$${target#*/}; suffix=; \
		if [ "$$os" = windows ]; then suffix=.exe; fi; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -trimpath -o "$(DIST)/tally-$$os-$$arch$$suffix" .; \
	done
