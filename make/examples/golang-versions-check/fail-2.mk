include $(addprefix ../../, \
	targets/golang/version.mk \
)

$(call verify-Dockerfile-builder-golang-version,images/Dockerfile-1.15)
$(call verify-go-mod-golang-version)

all: verify-golang-versions
	@echo "versions are correct"
.PHONY: all
