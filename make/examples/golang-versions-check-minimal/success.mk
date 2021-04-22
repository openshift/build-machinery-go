include $(addprefix ../../, \
	targets/golang/version.mk \
)

$(info makefile-test needs some output)
$(call verify-golang-versions,Dockerfile)

all: verify-golang-versions
	@echo "versions are correct"
.PHONY: all
