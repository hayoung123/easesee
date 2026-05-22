BIN_DIR ?= $(HOME)/.local/bin
SKILLS_DIR ?= $(HOME)/.claude/skills
REPO_DIR := $(shell pwd)

.PHONY: build install install-skills uninstall test clean

build:
	@mkdir -p bin
	go build -o bin/devs ./cmd/devs

install: build
	@mkdir -p $(BIN_DIR)
	install -m 0755 bin/devs $(BIN_DIR)/devs
	@echo "installed → $(BIN_DIR)/devs"

install-skills:
	@mkdir -p $(SKILLS_DIR)
	@for d in devs-register devs-help; do \
		rm -rf $(SKILLS_DIR)/$$d; \
		ln -s $(REPO_DIR)/skills/$$d $(SKILLS_DIR)/$$d; \
		echo "linked → $(SKILLS_DIR)/$$d"; \
	done

uninstall:
	rm -f $(BIN_DIR)/devs
	rm -rf $(SKILLS_DIR)/devs-register $(SKILLS_DIR)/devs-help
	@echo "removed binary and skill links (state and registry preserved)"

test:
	go test ./...

clean:
	rm -rf bin
