include $(addprefix ../../, \
	targets/golang/version.mk \
)

# Dockerfile has "AS builder" but no golang-X.Y in the image tag.
# The sed extraction returns empty, which must be rejected.
$(call verify-Dockerfile-builder-golang-version,images/Dockerfile-no-golang-version)
$(call verify-go-mod-golang-version)

all: verify-golang-versions
	@echo "versions are correct"
.PHONY: all
