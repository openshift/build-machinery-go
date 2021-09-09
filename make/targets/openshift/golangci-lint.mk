include $(addprefix $(dir $(lastword $(MAKEFILE_LIST))), \
	../../lib/golang.mk \
	../../lib/tmp.mk \
)

GOLANGCI_LINT_VERSION ?=1.42.1
GOLANGCI_LINT ?=$(PERMANENT_TMP_GOPATH)/bin/golangci-lint-$(GOLANGCI_LINT_VERSION)

golangci_lint_downloaded_filename := golangci-lint-$(GOLANGCI_LINT_VERSION)-$(GOHOSTOS)-$(GOHOSTARCH)
golangci_lint_dir := $(dir $(GOLANGCI_LINT))

.ensure-golangci-lint:
ifeq "" "$(wildcard $(GOLANGCI_LINT))"
	$(info Installing golangci-lint into '$(GOLANGCI_LINT)')
	mkdir -p '$(golangci_lint_dir)'
	curl -s -f -L https://github.com/golangci/golangci-lint/releases/download/v$(GOLANGCI_LINT_VERSION)/$(golangci_lint_downloaded_filename).tar.gz -o "$(PERMANENT_TMP_GOPATH)/bin/$(golangci_lint_downloaded_filename).tar.gz"
	tar -zxf "$(PERMANENT_TMP_GOPATH)/bin/$(golangci_lint_downloaded_filename).tar.gz" -C "$(PERMANENT_TMP_GOPATH)/bin"
	mv "$(PERMANENT_TMP_GOPATH)/bin/$(golangci_lint_downloaded_filename)/golangci-lint" $(GOLANGCI_LINT)
	rm -rf "$(PERMANENT_TMP_GOPATH)/bin/$(golangci_lint_downloaded_filename)"
	rm -rf "$(PERMANENT_TMP_GOPATH)/bin/$(golangci_lint_downloaded_filename).tar.gz"
	chmod +x '$(GOLANGCI_LINT)';
else
	$(info Using existing golangci-lint from $(GOLANGCI_LINT))
endif


ensure-controller-gen:
ifeq "" "$(wildcard $(CONTROLLER_GEN))"
	$(info Installing controller-gen into '$(CONTROLLER_GEN)')
	mkdir -p '$(controller_gen_dir)'
	curl -s -f -L https://github.com/openshift/kubernetes-sigs-controller-tools/releases/download/$(CONTROLLER_GEN_VERSION)/controller-gen-$(GOHOSTOS)-$(GOHOSTARCH) -o '$(CONTROLLER_GEN)'
	chmod +x '$(CONTROLLER_GEN)';
else
	$(info Using existing controller-gen from "$(CONTROLLER_GEN)")
	@[[ "$(_controller_gen_installed_version)" == $(CONTROLLER_GEN_VERSION) ]] || \
	echo "Warning: Installed controller-gen version $(_controller_gen_installed_version) does not match expected version $(CONTROLLER_GEN_VERSION)."
endif
.PHONY: ensure-controller-gen


verify-golangci-lint: .ensure-golangci-lint
	$(GOLANGCI_LINT) run \
	--timeout 30m \
	--disable-all \
	-E deadcode \
	-E unused \
	-E varcheck \
	-E ineffassign
.PHONY: verify-bindata

verify: verify-golangci-lint
.PHONY: verify