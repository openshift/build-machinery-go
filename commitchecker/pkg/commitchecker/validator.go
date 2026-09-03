package commitchecker

import "fmt"

type Validator interface {
	Name() string
	Validate(commit *Commit) []error
}

var (
	registry      = map[string]Validator{}
	registryOrder []string
)

func RegisterValidator(v Validator) {
	name := v.Name()
	if _, exists := registry[name]; exists {
		panic(fmt.Sprintf("duplicate validator registered: %q", name))
	}
	registry[name] = v
	registryOrder = append(registryOrder, name)
}

func ValidatorsForNames(names []string) ([]Validator, error) {
	// If no names are provided, return all built-in validators in the order they were registered
	if len(names) == 0 {
		var validators []Validator
		for _, name := range registryOrder {
			validators = append(validators, registry[name])
		}
		return validators, nil
	}

	// Otherwise, return the validators for the specified names in the order they were registered.
	var validators []Validator
	for _, name := range names {
		v, ok := registry[name]
		if !ok {
			return nil, fmt.Errorf("unknown validator %q; available validators: %v", name, registryOrder)
		}
		validators = append(validators, v)
	}
	return validators, nil
}

type builtinValidator struct {
	name string
	fn   func(*Commit) []error
}

func (b *builtinValidator) Name() string {
	return b.name
}

func (b *builtinValidator) Validate(commit *Commit) []error {
	return b.fn(commit)
}

func NewBuiltinValidator(name string, fn func(*Commit) []error) Validator {
	return &builtinValidator{name: name, fn: fn}
}
