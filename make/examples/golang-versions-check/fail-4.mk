include $(addprefix ../../, \
	targets/golang/version.mk \
)

$(call verify-golang-versions,images/Dockerfile-1.15)

all: verify-golang-versions
	@echo "versions are correct"
.PHONY: all
