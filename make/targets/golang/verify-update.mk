self_dir :=$(dir $(lastword $(MAKEFILE_LIST)))

go_files_count :=$(words $(GO_FILES))

verify-gofmt:
	$(info Running `$(GOFMT) $(GOFMT_FLAGS)` on $(go_files_count) file(s).)
	@TMP=$$( mktemp ); \
	$(GOFMT) $(GOFMT_FLAGS) $(GO_FILES) | tee $${TMP}; \
	if [ -s $${TMP} ]; then \
		echo "$@ failed - please run \`make update-gofmt\`"; \
		exit 1; \
	fi;
.PHONY: verify-gofmt

update-gofmt:
	$(info Running `$(GOFMT) $(GOFMT_FLAGS) -w` on $(go_files_count) file(s).)
	@$(GOFMT) $(GOFMT_FLAGS) -w $(GO_FILES)
.PHONY: update-gofmt


# FIXME: go vet needs to use $(GO_MOD_FLAGS) when this is fixed https://github.com/golang/go/issues/35955
# It will be enforced in CI by setting the env var there, so this remains to fix the dev experience
verify-govet:
	$(GO) vet $(GO_MOD_FLAGS) $(GO_PACKAGES)
.PHONY: verify-govet

verify-golint:
	$(GOLINT) $(GO_PACKAGES)
.PHONY: verify-govet

verify-gomod:
	$(info Running `$(GO) mod tidy`.)
	@if ! [ -f go.mod ]; then \
		echo "$@ failed - please run \`go mod init\`"; \
		exit 1; \
	fi; \
	if ! [ -f go.sum ]; then \
		echo "$@ failed - please run \`make update-gomod\`"; \
		exit 1; \
	fi; \
	TMP_GOMOD=$$( mktemp ); \
	TMP_GOSUM=$$( mktemp ); \
	cp --preserve=all go.mod $${TMP_GOMOD}; \
	cp --preserve=all go.sum $${TMP_GOSUM}; \
	$(GO) mod tidy; \
	if ! { diff $${TMP_GOMOD} go.mod && diff $${TMP_GOSUM} go.sum; }; then \
		mv $${TMP_GOMOD} go.mod; \
		mv $${TMP_GOSUM} go.sum; \
		echo "$@ failed - please run \`$(GO) mod tidy\`"; \
		exit 1; \
	fi;
.PHONY: verify-gomod

# We need to be careful to expand all the paths before any include is done
# or self_dir could be modified for the next include by the included file.
# Also doing this at the end of the file allows us to use self_dir before it could be modified.
include $(addprefix $(self_dir), \
	../../lib/golang.mk \
)
