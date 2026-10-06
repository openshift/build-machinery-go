# library-go/build-machinery-go
These are the building blocks for this and many of our other repositories to share code for Makefiles, helper scripts and other build related machinery.

## Makefiles
`make/` directory contains several predefined makefiles `(*.mk)` to choose from and include one of them as a base in your final `Makefile`. These are the predefined flows providing you with e.g. `build`, `test` or `verify` targets. To start with it is recommended you base Makefile on the corresponding `*.example.mk` using copy&paste.

As some advanced targets are generated, every Makefile contains `make help` target listing all the available ones. All of the "example" makefiles have a corresponding `.help` file listing all the targets available there.

Also for advanced use and if none of the predefined flows doesn't fit your needs, you can compose the flow from modules in similar way to how the predefined flows do,  

### Golang
Standard makefile for building pure Golang projects.
 - [make/golang.mk](make/golang.mk)
 - [make/golang.example.mk](make/golang.example.mk)
 - [make/golang.example.mk.help](make/golang.example.mk.help)

### Default
Standard makefile for OpenShift Golang projects. 

Extends [#Golang]().

 - [make/default.mk](make/default.mk)
 - [make/default.example.mk](make/default.example.mk)
 - [make/default.example.mk.help](make/default.example.mk.help)

### Operator
Standard makefile for OpenShift Golang projects. 

Extends [#Default]().

 - [make/operator.mk](make/operator.mk)
 - [make/operator.example.mk](make/operator.example.mk)
 - [make/operator.example.mk.help](make/operator.example.mk.help)


## Scripts
`scripts` contain more complicated logic that is used in some make targets.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development workflow, PR guidelines,
and how to regenerate checked-in makefile logs.

For architecture details, see [ARCHITECTURE.md](ARCHITECTURE.md). For AI agent
instructions, see [AGENTS.md](AGENTS.md) ([CLAUDE.md](CLAUDE.md) references it via `@AGENTS.md`).

## Optional crypto source inventory

`make crypto-inventory` invokes a pinned `check-payload` supplied by the build
root and writes `_output/crypto-inventory/source.json`. It is opt-in and does not
change `make verify`. Select only executable packages shipped in the component's
images, and run in the production build environment:

```sh
make crypto-inventory CRYPTO_SCANNER=/tools/check-payload \
  CRYPTO_SOURCE_PACKAGES='./cmd/server ./cmd/helper' \
  CRYPTO_COMPONENT=example CRYPTO_CGO_ENABLED=1
```

The target passes `GOOS`, `GOARCH`, `GO_MOD_FLAGS` and `GO_BUILD_FLAGS`. Override
`CRYPTO_BUILD_FLAGS` if the production build uses additional tags. Analysis lives
in check-payload; this repository provides invocation only. Test executables and
custom Kubernetes build scripts can invoke the scanner directly. Findings are
provisional classifications, and incomplete analysis fails the target while
preserving the JSON report. CI must publish and schema-validate the report, and
collect shipped-binary evidence separately. See the check-payload crypto inventory
documentation for limits and rollout guidance.
