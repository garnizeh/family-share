# Makefile - repo helper targets
#
# Usage:
#   make help
#   make clean-local-branches-dry   # show branches that would be deleted
#   make clean-local-branches       # delete local branches merged into origin/main
#   make clean-local-branches-force # force-delete local branches merged into origin/main

.PHONY: help clean-local-branches clean-local-branches-dry clean-local-branches-force

DEFAULT_BRANCH ?= main
REMOTE ?= origin

help:
	@echo "Makefile helper targets"
	@echo
	@echo "  clean-local-branches-dry   - show local branches merged into $(REMOTE)/$(DEFAULT_BRANCH) (dry run)"
	@echo "  clean-local-branches       - delete local branches merged into $(REMOTE)/$(DEFAULT_BRANCH) (safe delete -\d)"
	@echo "  clean-local-branches-force - force delete local branches merged into $(REMOTE)/$(DEFAULT_BRANCH) (\-D)"

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
