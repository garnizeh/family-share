# Makefile - repo helper targets
#
# Usage:
#   make help
#   make clean-local-branches-dry   # show branches that would be deleted
#   make clean-local-branches       # delete local branches merged into origin/main
#   make clean-local-branches-force # force-delete local branches merged into origin/main

.PHONY: help clean-local-branches clean-local-branches-dry clean-local-branches-force build run-local

DEFAULT_BRANCH ?= main
REMOTE ?= origin

help:
	@echo "Makefile helper targets"
	@echo
	@echo "  clean-local-branches-dry   - show local branches merged into $(REMOTE)/$(DEFAULT_BRANCH) (dry run)"
	@echo "  clean-local-branches       - delete local branches merged into $(REMOTE)/$(DEFAULT_BRANCH) (safe delete -\d)"
	@echo "  clean-local-branches-force - force-delete local branches merged into $(REMOTE)/$(DEFAULT_BRANCH) (\-D)"
	@echo "  build                      - compile binary to bin/familyshare"
	@echo "  run-local                  - build and run locally (PORT, TEMP_UPLOAD_DIR env vars supported)"

# Build and run helpers for local testing
# Usage:
#   make build         # compile binary to bin/familyshare
#   make run-local     # build then run with sensible defaults (PORT=8080, TEMP_UPLOAD_DIR=./tmp_uploads)

BINARY ?= bin/familyshare
MAIN_PKG ?= ./cmd/app
PORT ?= 8080
TEMP_UPLOAD_DIR ?= $(CURDIR)/tmp_uploads

build:
	@echo "Building $(BINARY) from $(MAIN_PKG)..."
	@mkdir -p $(dir $(BINARY))
	@go build -o $(BINARY) $(MAIN_PKG)

run-local: build
	@echo "Preparing local temp upload dir: $(TEMP_UPLOAD_DIR)"
	@mkdir -p $(TEMP_UPLOAD_DIR)
	@PORT=$(PORT) TEMP_UPLOAD_DIR=$(TEMP_UPLOAD_DIR) $(BINARY)

# Dry run: list branches merged into remote default branch, exclude protected names
clean-local-branches-dry:
	@echo "Fetching latest refs from $(REMOTE)..."
	@git fetch --prune $(REMOTE)
	@echo
	@echo "Local branches merged into $(REMOTE)/$(DEFAULT_BRANCH) (dry run):"
	@git branch --merged $(REMOTE)/$(DEFAULT_BRANCH) | \
		grep -vE "(^\*|\b$(DEFAULT_BRANCH)\b|\bmaster\b|\bdevelop\b)" | sed 's/^..//' || true

# Safe delete: use git branch -d which refuses to delete unmerged branches
clean-local-branches:
	@echo "Fetching latest refs from $(REMOTE)..."
	@git fetch --prune $(REMOTE)
	@for b in `git branch --merged $(REMOTE)/$(DEFAULT_BRANCH) | grep -vE "(^\*|\b$(DEFAULT_BRANCH)\b|\bmaster\b|\bdevelop\b)" | sed 's/^..//'`; do \
		printf "Deleting local branch '%s'\n" "$$b"; \
		git branch -d "$$b"; \
	done

# Force delete: use git branch -D
clean-local-branches-force:
	@echo "Fetching latest refs from $(REMOTE)..."
	@git fetch --prune $(REMOTE)
	@for b in `git branch --merged $(REMOTE)/$(DEFAULT_BRANCH) | grep -vE "(^\*|\b$(DEFAULT_BRANCH)\b|\bmaster\b|\bdevelop\b)" | sed 's/^..//'`; do \
		printf "Force deleting local branch '%s'\n" "$$b"; \
		git branch -D "$$b"; \
	done
