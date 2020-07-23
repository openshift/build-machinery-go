help: ## Print help for available make targets.
	$(info The following make targets are available:)
	@sed -n 's/^\([a-zA-Z_\-]*\): .*## \(.*\)/  \1\t\2/p' $(MAKEFILE_LIST) | sort
.PHONY: help
