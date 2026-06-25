include $(addprefix ../../, \
	targets/golang/version.mk \
)

# Finding 2: Full ART scenario with all three sources.
# go.mod (1.16) < Dockerfile (1.17) + .ci-operator.yaml (1.17)
# success-4.mk only tests two sources (Dockerfile + go.mod).
$(call verify-Dockerfile-builder-golang-version,images/Dockerfile-1.17)
$(call verify-golang-version-reference,.ci-operator.yaml,1.17)
$(call verify-go-mod-golang-version)

all: verify-golang-versions
	@echo "versions are correct"
.PHONY: all
