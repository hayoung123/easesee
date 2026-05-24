BIN_DIR    ?= $(HOME)/.local/bin
SKILLS_DIR ?= $(HOME)/.claude/skills
REPO_DIR   := $(shell pwd)

# VERSION can be overridden: `make release-binaries VERSION=0.2.0`
# When unset, defaults to the closest git tag (or "dev").
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null | sed 's/^v//' || echo dev)

LDFLAGS := -s -w -X 'github.com/hayoung123/easesee/internal/cli.Version=$(VERSION)'

PLATFORMS := darwin-arm64 darwin-amd64 linux-amd64 linux-arm64

.PHONY: build install install-skills uninstall test clean release-binaries

build:
	@mkdir -p bin
	go build -ldflags="$(LDFLAGS)" -o bin/easesee ./cmd/easesee

install: build
	@mkdir -p $(BIN_DIR)
	install -m 0755 bin/easesee $(BIN_DIR)/easesee
	@echo "installed → $(BIN_DIR)/easesee (version $(VERSION))"

install-skills:
	@mkdir -p $(SKILLS_DIR)
	@for d in easesee-register easesee-help; do \
		rm -rf $(SKILLS_DIR)/$$d; \
		ln -s $(REPO_DIR)/skills/$$d $(SKILLS_DIR)/$$d; \
		echo "linked → $(SKILLS_DIR)/$$d"; \
	done

uninstall:
	rm -f $(BIN_DIR)/easesee
	rm -rf $(SKILLS_DIR)/easesee-register $(SKILLS_DIR)/easesee-help
	@echo "removed binary and skill links (state and registry preserved)"

test:
	go test ./...

# Cross-compiles release binaries into ./dist with VERSION baked in.
# Usage: make release-binaries VERSION=0.2.0
release-binaries:
	@mkdir -p dist
	@for p in $(PLATFORMS); do \
		os=$${p%-*}; arch=$${p#*-}; \
		echo "building dist/easesee-$$os-$$arch (v$(VERSION))"; \
		GOOS=$$os GOARCH=$$arch go build -ldflags="$(LDFLAGS)" -o dist/easesee-$$os-$$arch ./cmd/easesee; \
	done

publish:
	cd npm && npm publish --access public

clean:
	rm -rf bin dist
