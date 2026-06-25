include $(addprefix ../../, \
	targets/golang/version.mk \
)

# Finding 1: Dockerfile with no extractable golang-MAJOR.MINOR pattern.
# The sed regex returns the entire FROM line, producing non-numeric
# version components. The arithmetic comparison silently fails under
# 2>/dev/null, so the check passes when it should fail.
$(call verify-Dockerfile-builder-golang-version,images/Dockerfile-malformed)
$(call verify-go-mod-golang-version)

all: verify-golang-versions
	@echo "versions are correct"
.PHONY: all
