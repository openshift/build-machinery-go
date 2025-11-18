# OpenShift Test Suite targets
#
# This file provides common targets for building and managing OpenShift test suites
# across all operator repos that use the openshift-tests-extension framework.
#
# Usage in main Makefile:
#   include $(addprefix ./vendor/github.com/openshift/build-machinery-go/make/, \
#       targets/openshift/test-suite.mk \
#   )
#
#   TESTS_BINARY := my-operator-tests
#   TESTS_DIR := ./cmd/my-operator-tests
#
# Required variables:
#   TESTS_BINARY - Name of the test binary
#   TESTS_DIR - Directory containing test code (e.g., ./cmd/my-operator-tests)

# -------------------------------------------------------------------
# Build and move test binary to correct location
# -------------------------------------------------------------------
.PHONY: tests-ext-build
tests-ext-build: build
	@mkdir -p $(TESTS_DIR)
	@if [ -f $(shell basename $(TESTS_DIR)) ] && [ ! -f $(TESTS_DIR)/$(TESTS_BINARY) ]; then \
		mv $(shell basename $(TESTS_DIR)) $(TESTS_DIR)/$(TESTS_BINARY); \
	fi

