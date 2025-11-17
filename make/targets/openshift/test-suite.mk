# OpenShift Test Suite targets
#
# This file provides common targets for running OpenShift test suites
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
# Run test suite
# -------------------------------------------------------------------
.PHONY: run-suite
run-suite:
	@if [ -z "$(SUITE)" ]; then \
		echo "Error: SUITE variable is required. Usage: make run-suite SUITE=<suite-name> [JUNIT_DIR=<dir>]"; \
		exit 1; \
	fi
	@JUNIT_ARG=""; \
	if [ -n "$(JUNIT_DIR)" ]; then \
		mkdir -p "$(JUNIT_DIR)"; \
		JUNIT_ARG="--junit-path=$(JUNIT_DIR)/junit.xml"; \
	fi; \
	"$(TESTS_DIR)/$(TESTS_BINARY)" run-suite $(SUITE) $$JUNIT_ARG

