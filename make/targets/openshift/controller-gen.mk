include $(addprefix $(dir $(lastword $(MAKEFILE_LIST))), \
	../../lib/golang.mk \
	../../lib/tmp.mk \
)

CONTROLLER_GEN_VERSION ?=v0.6.0
CONTROLLER_GEN ?=$(PERMANENT_TMP_GOPATH)/bin/controller-gen
controller_gen_dir :=$(dir $(CONTROLLER_GEN))

ensure-controller-gen:
ifeq "" "$(wildcard $(CONTROLLER_GEN))"
	$(info Installing controller-gen into '$(CONTROLLER_GEN)')
	mkdir -p '$(controller_gen_dir)'
	curl -s -f -L https://github.com/openshift/kubernetes-sigs-controller-tools/releases/download/$(CONTROLLER_GEN_VERSION)/controller-gen-$(GOHOSTOS)-$(GOHOSTARCH) -o '$(CONTROLLER_GEN)'
	chmod +x '$(CONTROLLER_GEN)';
else
	$(info Using existing controller-gen from "$(CONTROLLER_GEN)")
endif
.PHONY: ensure-controller-gen

clean-controller-gen:
	$(RM) '$(CONTROLLER_GEN)'
	if [ -d '$(controller_gen_dir)' ]; then rmdir --ignore-fail-on-non-empty -p '$(controller_gen_dir)'; fi
.PHONY: clean-controller-gen

clean: clean-controller-gen
