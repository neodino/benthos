package metrics

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/benthosdev/benthos/v4/internal/docs"
)

// Config is the all encompassing configuration struct for all metric output
// types.
type Config struct {
	Type    string `json:"type" yaml:"type"`
	Mapping string `json:"mapping" yaml:"mapping"`
	Plugin  any    `json:"plugin,omitempty" yaml:"plugin,omitempty"`
}

// NewConfig returns a configuration struct fully populated with default values.
func NewConfig() Config {
	return Config{
		Type:    docs.DefaultTypeOf(docs.TypeMetrics),
		Mapping: "",
		Plugin:  nil,
	}
}

//------------------------------------------------------------------------------

// UnmarshalYAML ensures that when parsing configs that are in a map or slice
// the default values are still applied.
func (conf *Config) UnmarshalYAML(value *yaml.Node) error {
	type confAlias Config
	aliased := confAlias(NewConfig())

	err := value.Decode(&aliased)
	if err != nil {
		return docs.NewLintError(value.Line, docs.LintFailedRead, err)
	}

	var spec docs.ComponentSpec
	if aliased.Type, spec, err = docs.GetInferenceCandidateFromYAML(docs.DeprecatedProvider, docs.TypeMetrics, value); err != nil {
		return docs.NewLintError(value.Line, docs.LintComponentMissing, err)
	}

	if spec.Plugin {
		pluginNode, err := docs.GetPluginConfigYAML(aliased.Type, value)
		if err != nil {
			return docs.NewLintError(value.Line, docs.LintFailedRead, err)
		}
		aliased.Plugin = &pluginNode
	} else {
		aliased.Plugin = nil
	}

	*conf = Config(aliased)
	return nil
}

func FromAny(prov docs.Provider, value any) (conf Config, err error) {
	switch t := value.(type) {
	case Config:
		return t, nil
	case *yaml.Node:
		return fromYAML(prov, t)
	case map[string]any:
		return fromMap(prov, t)
	}
	err = fmt.Errorf("unexpected value, expected object, got %T", value)
	return
}

func fromMap(prov docs.Provider, value map[string]any) (conf Config, err error) {
	if conf.Type, _, err = docs.GetInferenceCandidateFromMap(prov, docs.TypeMetrics, value); err != nil {
		err = docs.NewLintError(0, docs.LintComponentNotFound, err)
		return
	}

	if p, exists := value[conf.Type]; exists {
		conf.Plugin = p
	} else if p, exists := value["plugin"]; exists {
		conf.Plugin = p
	}
	return
}

func fromYAML(prov docs.Provider, value *yaml.Node) (conf Config, err error) {
	if conf.Type, _, err = docs.GetInferenceCandidateFromYAML(prov, docs.TypeMetrics, value); err != nil {
		err = docs.NewLintError(value.Line, docs.LintComponentNotFound, err)
		return
	}

	pluginNode, err := docs.GetPluginConfigYAML(conf.Type, value)
	if err != nil {
		err = docs.NewLintError(value.Line, docs.LintFailedRead, err)
		return
	}

	conf.Plugin = &pluginNode
	return
}
