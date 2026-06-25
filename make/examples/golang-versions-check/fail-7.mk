include $(addprefix ../../, \
	targets/golang/version.mk \
)

# Simulates what happens when a version extraction returns empty
# (e.g., .ci-operator.yaml has build_root_image but no golang-X.Y tag).
$(call verify-golang-version-reference,simulated-empty-source,)
$(call verify-go-mod-golang-version)

all: verify-golang-versions
	@echo "versions are correct"
.PHONY: all
