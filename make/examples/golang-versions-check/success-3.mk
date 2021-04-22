include $(addprefix ../../, \
	targets/golang/version.mk \
)

$(call verify-golang-versions)

all: verify-golang-versions
	@echo "versions are correct"
.PHONY: all
