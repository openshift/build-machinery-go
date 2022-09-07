include $(addprefix $(dir $(lastword $(MAKEFILE_LIST))), \
	../../lib/golang.mk \
	../../lib/tmp.mk \
	../../targets/openshift/controller-gen.mk \
	../../targets/openshift/yq.mk \
	../../targets/openshift/yaml-patch.mk \
)

# $1 - crd file
# $2 - patch file
define patch-crd-yq
	$(YQ) m -i -x '$(1)' '$(2)'

endef

# $1 - crd file
# $2 - patch file
define patch-crd-yaml-patch
	env | grep OPENSHIFT
	$(YAML_PATCH) -o '$(2)' < '$(1)' > '$(1).patched'
    mv '$(1).patched' '$(1)'

endef

empty :=

# $1 - apis
# $2 - manifests
define run-crd-gen
env | grep OPENSHIFT
ifndef OPENSHIFT_REQUIRED_FEATURESET
	@echo "MISSING!"
else
	@echo "PRESENT"
endif
	'$(CONTROLLER_GEN)' \
		schemapatch:manifests="$(2)" \
		paths="$(subst $(empty) ,;,$(1))" \
		'output:dir="$(2)"'
	$$(foreach p,$$(wildcard $(2)/*.crd.yaml-merge-patch),$$(call patch-crd-yq,$$(basename $$(p)).yaml,$$(p)))
	$$(foreach p,$$(wildcard $(2)/*.crd.yaml-patch),$$(call patch-crd-yaml-patch,$$(basename $$(p)).yaml,$$(p)))
endef


# $1 - target name
# $2 - apis
# $3 - manifests
define add-crd-gen-internal

update-codegen-crds-$(1): ensure-controller-gen ensure-yq ensure-yaml-patch
	$(call run-crd-gen,$(2),$(3))
.PHONY: update-codegen-crds-$(1)

update-codegen-crds: update-codegen-crds-$(1)
.PHONY: update-codegen-crds

verify-codegen-crds-$(1): update-codegen-crds-$(1)
	git diff --exit-code
.PHONY: verify-codegen-crds-$(1)

verify-codegen-crds: verify-codegen-crds-$(1)
.PHONY: verify-codegen-crds

update-codegen-techpreview-crds-$(1): ensure-controller-gen ensure-yq ensure-yaml-patch
	OPENSHIFT_REQUIRED_FEATURESET=TechPreviewNoUpgrade $(call run-crd-gen,$(2),$(3))
.PHONY: update-codegen-techpreview-crds-$(1)

update-codegen-techpreview-crds: update-codegen-techpreview-crds-$(1)
.PHONY: update-codegen-techpreview-crds

verify-codegen-techpreview-crds-$(1): update-codegen-techpreview-crds-$(1)
	git diff --exit-code
.PHONY: verify-codegen-techpreview-crds-$(1)

verify-codegen-techpreview-crds: verify-codegen-techpreview-crds-$(1)
.PHONY: verify-codegen-techpreview-crds

update-codegen-default-crds-$(1): ensure-controller-gen ensure-yq ensure-yaml-patch
	OPENSHIFT_REQUIRED_FEATURESET=Default $(call run-crd-gen,$(2),$(3))
.PHONY: update-codegen-default-crds-$(1)

update-codegen-default-crds: update-codegen-default-crds-$(1)
.PHONY: update-codegen-default-crds

verify-codegen-default-crds-$(1): update-codegen-default-crds-$(1)
	git diff --exit-code
.PHONY: verify-codegen-default-crds-$(1)

verify-codegen-default-crds: verify-codegen-default-crds-$(1)
.PHONY: verify-codegen-default-crds

update-codegen-customNoUpgrade-crds-$(1): ensure-controller-gen ensure-yq ensure-yaml-patch
	OPENSHIFT_REQUIRED_FEATURESET=CustomNoUpgrade $(call run-crd-gen,$(2),$(3))
.PHONY: update-codegen-crds-$(1)

update-codegen-customNoUpgrade-crds: update-codegen-customNoUpgrade-crds-$(1)
.PHONY: update-codegen-customNoUpgrade-crds

verify-codegen-customNoUpgrade-crds-$(1): update-codegen-customNoUpgrade-crds-$(1)
	git diff --exit-code
.PHONY: verify-codegen-customNoUpgrade-crds-$(1)

verify-codegen-customNoUpgrade-crds: verify-codegen-customNoUpgrade-crds-$(1)
.PHONY: verify-codegen-customNoUpgrade-crds

endef


update-generated: update-codegen-crds update-codegen-techpreview-crds update-codegen-default-crds update-codegen-customNoUpgrade-crds
.PHONY: update-generated

update: update-generated
.PHONY: update

verify-generated: verify-codegen-crds verify-codegen-techpreview-crds verify-codegen-default-crds verify-codegen-customNoUpgrade-crds
.PHONY: verify-generated

verify: verify-generated
.PHONY: verify


define add-crd-gen
$(eval $(call add-crd-gen-internal,$(1),$(2),$(3)))
endef

