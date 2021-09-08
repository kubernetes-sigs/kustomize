package types

import (
	"reflect"
	"strings"

	yaml2json "sigs.k8s.io/yaml"

	"sigs.k8s.io/kustomize/kyaml/errors"
	"sigs.k8s.io/kustomize/kyaml/fn/runtime/runtimeutil"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

const (
	CompositionVersion = "kustomize.config.k8s.io/v1alpha1"
	CompositionKind    = "Composition"
)

// Composition is a client-side Kustomize resource for composing sets of transformers.
// Compositions may either be built into resource collections directly with kustomize build,
// or be composed into other Compositions using transformersFrom. When a Composition contains multiple transformers,
// the output of each transformer is provided as input to the next transformer.
type Composition struct {
	TypeMeta `json:",inline" yaml:",inline"`

	// TransformersFrom lists Compositions to import via relative paths, absolute paths, or (in the future) URLs.
	// Imported transformers are merged with inline transformers from the `transformers` field before execution.
	TransformersFrom []TransformerSource `json:"transformersFrom,omitempty" yaml:"transformersFrom,omitempty"`

	// Transformers is the list of transformers to invoke in order when executing this Composition.
	Transformers []Transformer `json:"transformers,omitempty" yaml:"transformers,omitempty"`

	// TransformerOverrides is a list of patches to be applied to imported transformers by the same name
	TransformerOverrides []Transformer `json:"transformerOverrides,omitempty" yaml:"transformerOverrides,omitempty"`

	// TransformerOrder is a list of transformer identifiers used to override the ordering of the consolidated transformers.
	// If specified, it must include all transformers in the consolidated list.
	// Transformers may be referred to by name if the name is unique. Otherwise, transformer GVK must also be specified.
	TransformerOrder []resid.ResId `json:"transformerOrder,omitempty" yaml:"transformerOrder,omitempty"`
}

func (c *Composition) Default() error {
	for i := range c.Transformers {
		if err := c.Transformers[i].Default(); err != nil {
			return err
		}
	}
	for i := range c.TransformerOverrides {
		if err := c.TransformerOverrides[i].Default(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Composition) Validate() error {
	if len(c.TransformersFrom) == 0 && len(c.TransformerOverrides) > 0 {
		return errors.Errorf("%s contained overrides but did not import any transformers", CompositionKind)
	}
	for _, transformerSource := range c.TransformersFrom {
		if transformerSource.Path == "" {
			return errors.Errorf("transformersFrom entry is missing the path field")
		}
	}
	return nil
}

// Transformer is a composable unit that generates, transforms and/or validates a collection of resources in a Composition.
// Transformers are defined as client-side resources with standard Kubernetes metadata.
// Each transformer takes a collection of resources as input, and emits a collection of resources as output.
// The output of a transformer replaces the previous resource collection – i.e. transformers form a pipeline of transformations.
type Transformer struct {
	TypeMeta `json:",inline" yaml:",inline"`

	// MetaData is a pointer to avoid marshalling empty struct
	MetaData ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	// RuntimeConfig specifies where to find the implementation of the transformer, i.e. the code to invoke to execute it.
	RuntimeConfig runtimeutil.FunctionSpec `yaml:"runtime,omitempty" json:"runtime,omitempty"`

	// TransformerConfig contains the transformer options that will be provided during transformer execution.
	TransformerConfig yaml.Node `yaml:",inline" json:",inline"`
}

// TransformerSource contains the configuration for importing a set of transformers from another Composition.
type TransformerSource struct {
	// Path is the relative path to the Composition you want to import.
	Path string `json:"path,omitempty" yaml:"path,omitempty"`

	// ImportMode defines how to merge the target Composition's transformers with the current ones.
	// Accepted values are 'prepend' (default) or 'append'.
	ImportMode string `json:"importMode,omitempty" yaml:"importMode,omitempty"`
}

func (t *Transformer) ID() resid.ResId {
	group, version := resid.ParseGroupVersion(t.APIVersion)
	return resid.ResId{
		Gvk: resid.Gvk{
			Group:   group,
			Version: version,
			Kind:    t.Kind,
		},
		Name:      t.MetaData.Name,
		Namespace: t.MetaData.Namespace,
	}
}

func (t Transformer) String() string {
	return t.ID().String()
}

func (t *Transformer) AsYAML() (string, error) {
	b, err := yaml.Marshal(t)
	return string(b), err
}

// UnmarshalYAML makes Transformer implement yaml.Unmarshaler.
// This is a workaround this bug: https://github.com/go-yaml/yaml/issues/672
func (t *Transformer) UnmarshalYAML(node *yaml.Node) error {
	// We clone the incoming node rather than creating a new one from scratch to keep
	// the node's metadata, including kind and tag but also column and line numbers.
	t.TransformerConfig = *node
	t.TransformerConfig.Content = nil

	for i := 0; i < len(node.Content); i += 2 {
		var dest interface{}
		switch key := node.Content[i]; key.Value {
		case "apiVersion":
			dest = &t.APIVersion
		case "kind":
			dest = &t.Kind
		case "metadata":
			dest = &t.MetaData
		case "runtime":
			dest = &t.RuntimeConfig
		default:
			t.TransformerConfig.Content = append(t.TransformerConfig.Content, node.Content[i], node.Content[i+1])
			continue
		}

		value := node.Content[i+1]
		if err := value.Decode(dest); err != nil {
			return err
		}
	}
	// TODO: decide how to handle runtime config and remove this hack
	configAnno, annoErr := yaml.Marshal(t.RuntimeConfig)
	if annoErr != nil {
		return errors.WrapPrefixf(annoErr, "failed to marshal runtime config to annotations")
	}
	if t.MetaData.Annotations == nil {
		t.MetaData.Annotations = map[string]string{}
	}
	t.MetaData.Annotations[runtimeutil.FunctionAnnotationKey] = string(configAnno)
	return nil
}

// MarshalYAML makes Transformer implement yaml.Marshaler.
// This is a workaround this bug: https://github.com/go-yaml/yaml/issues/672
func (t Transformer) MarshalYAML() (interface{}, error) { return t.toYAML() }

// toYAML returns the YAML form of the Transformer in the form expected by transformers for execution.
func (t Transformer) toYAML() (*yaml.Node, error) {
	// TODO: decide how to handle runtime config and remove this hack
	configAnno, annoErr := yaml.Marshal(t.RuntimeConfig)
	if annoErr != nil {
		return nil, errors.WrapPrefixf(annoErr, "failed to marshal runtime config to annotations")
	}
	t.MetaData.Annotations[runtimeutil.FunctionAnnotationKey] = string(configAnno)

	// We clone the incoming node rather than creating a new one from scratch to keep
	// the node's metadata, including kind and tag but also column and line numbers.
	node := t.TransformerConfig
	node.Content = nil

	props := []struct {
		name  string
		value interface{}
	}{
		{"apiVersion", t.APIVersion},
		{"kind", t.Kind},
		{"metadata", t.MetaData},
		{"runtime", t.RuntimeConfig},
	}

	for _, prop := range props {
		if reflect.ValueOf(prop.value).IsZero() {
			continue
		}
		var name, value yaml.Node
		if err := name.Encode(prop.name); err != nil {
			return nil, err
		}
		if err := value.Encode(prop.value); err != nil {
			return nil, err
		}
		node.Content = append(node.Content, &name, &value)
	}

	node.Content = append(node.Content, t.TransformerConfig.Content...)
	return &node, nil
}

// MarshalJSON makes Transformer satisfy the json.Marshaler interface.
func (t Transformer) MarshalJSON() ([]byte, error) {
	// It is necessary to encode to YAML and then JSON to avoid issues with struct embedded fields.
	bs, err := yaml.Marshal(t)
	if err != nil {
		return nil, errors.WrapPrefixf(err, "could not marshal transfomer to YAML")
	}
	return yaml2json.YAMLToJSONStrict(bs)
}

func (t *Transformer) Default() error {
	if t.MetaData.Name == "" {
		t.MetaData.Name = strings.ToLower(t.Kind)
	}
	if t.MetaData.Namespace == "" {
		t.MetaData.Namespace = "default"
	}
	if t.APIVersion == "" {
		t.APIVersion = "builtin"
	}
	return nil
}
