include $(addprefix $(dir $(lastword $(MAKEFILE_LIST))), \
	../../lib/golang.mk \
)

# The CI build root supplies a pinned scanner binary. This target never installs
# a floating version and remains opt-in rather than changing `make verify`.
CRYPTO_SCANNER ?=check-payload
CRYPTO_SOURCE_PACKAGES ?=$(GO_BUILD_PACKAGES)
CRYPTO_BUILD_FLAGS ?=$(GO_MOD_FLAGS) $(GO_BUILD_FLAGS)
CRYPTO_COMPONENT ?=$(GO_PACKAGE)
CRYPTO_INVENTORY_DIR ?=_output/crypto-inventory
CRYPTO_INVENTORY_FILE ?=$(CRYPTO_INVENTORY_DIR)/source.json
CRYPTO_CGO_ENABLED ?=$(shell $(GO) env CGO_ENABLED)

crypto_empty :=
crypto_space :=$(crypto_empty) $(crypto_empty)
crypto_comma :=,
crypto_packages :=$(subst $(crypto_space),$(crypto_comma),$(strip $(CRYPTO_SOURCE_PACKAGES)))

crypto-inventory:
	$(if $(strip $(crypto_packages)),,$(error CRYPTO_SOURCE_PACKAGES must select shipped executable packages))
	mkdir -p '$(dir $(CRYPTO_INVENTORY_FILE))'
	'$(CRYPTO_SCANNER)' inventory source \
		--dir . --packages '$(crypto_packages)' --component '$(CRYPTO_COMPONENT)' \
		--goos '$(GOOS)' --goarch '$(GOARCH)' --cgo-enabled '$(CRYPTO_CGO_ENABLED)' \
		$(foreach flag,$(CRYPTO_BUILD_FLAGS),--build-flag='$(flag)') \
		--output-file '$(CRYPTO_INVENTORY_FILE)'
.PHONY: crypto-inventory
