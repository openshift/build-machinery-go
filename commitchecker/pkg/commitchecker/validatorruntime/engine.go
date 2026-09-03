package validatorruntime

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/openshift/build-machinery-go/commitchecker/pkg/commitchecker"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) LoadValidator(name string, src []byte) (commitchecker.Validator, error) {
	i := interp.New(interp.Options{})

	if err := i.Use(filteredStdlib()); err != nil {
		return nil, fmt.Errorf("failed to load stdlib: %w", err)
	}

	if err := i.Use(commitcheckerSymbols()); err != nil {
		return nil, fmt.Errorf("failed to load commitchecker symbols: %w", err)
	}

	_, err := i.Eval(string(src))
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate validator %q: %w", name, err)
	}

	pkgName := extractPackageName(string(src))

	v, err := i.Eval(pkgName + ".Validate")
	if err != nil {
		return nil, fmt.Errorf("validator %q must export a Validate function: %w", name, err)
	}

	fn, ok := v.Interface().(func(*commitchecker.Commit) []error)
	if !ok {
		return nil, fmt.Errorf("validator %q Validate function has wrong signature, expected func(*commitchecker.Commit) []error", name)
	}

	return &dynamicValidator{name: name, fn: fn}, nil
}

func extractPackageName(src string) string {
	for _, line := range strings.Split(src, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "package ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "package"))
		}
	}
	return "main"
}

var allowedStdlibPrefixes = []string{
	"fmt/",
	"strings/",
	"regexp/",
	"path/",
	"strconv/",
	"sort/",
	"math/",
	"unicode/",
	"bytes/",
	"errors/",
}

func filteredStdlib() interp.Exports {
	filtered := make(interp.Exports)
	for k, v := range stdlib.Symbols {
		if k == "." {
			filtered[k] = v
			continue
		}
		for _, prefix := range allowedStdlibPrefixes {
			if strings.HasPrefix(k, prefix) {
				filtered[k] = v
				break
			}
		}
	}
	return filtered
}

func commitcheckerSymbols() interp.Exports {
	return interp.Exports{
		"github.com/openshift/build-machinery-go/commitchecker/pkg/commitchecker/commitchecker": {
			"Commit": reflect.ValueOf((*commitchecker.Commit)(nil)),
		},
	}
}

type dynamicValidator struct {
	name string
	fn   func(*commitchecker.Commit) []error
}

func (d *dynamicValidator) Name() string {
	return d.name
}

func (d *dynamicValidator) Validate(commit *commitchecker.Commit) []error {
	return d.fn(commit)
}
