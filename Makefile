BIN_DIR ?= $(HOME)/.local/bin
SKILLS_DIR ?= $(HOME)/.claude/skills
REPO_DIR := $(shell pwd)

.PHONY: build install install-skills uninstall test clean

build:
	@mkdir -p bin
	go build -o bin/easesee ./cmd/easesee

install: build
	@mkdir -p $(BIN_DIR)
	install -m 0755 bin/easesee $(BIN_DIR)/easesee
	@echo "installed → $(BIN_DIR)/easesee"

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

clean:
	rm -rf bin
