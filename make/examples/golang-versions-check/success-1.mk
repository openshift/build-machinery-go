include $(addprefix ../../, \
	targets/golang/version.mk \
)

$(call verify-Dockerfile-builder-golang-version,images/Dockerfile-1.16)
$(call verify-go-mod-golang-version)
$(call verify-buildroot-golang-version)

all: verify-golang-versions
	@echo "versions are correct"
.PHONY: all
