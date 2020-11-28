// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package generators

import (
	"encoding/base64"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/go-errors/errors"
	"sigs.k8s.io/kustomize/api/filters/filtersutil"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func makeBaseNode(kind, name, namespace string) (*yaml.RNode, error) {
	rn, err := yaml.Parse(fmt.Sprintf(`
apiVersion: v1
kind: %s
`, kind))
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, errors.Errorf("a configmap must have a name")
	}
	if _, err := rn.Pipe(yaml.SetK8sName(name)); err != nil {
		return nil, err
	}
	if namespace != "" {
		if _, err := rn.Pipe(yaml.SetK8sNamespace(namespace)); err != nil {
			return nil, err
		}
	}
	return rn, nil
}

func makeValidatedDataMap(
	ldr ifc.KvLoader, name string, sources types.KvPairSources) (map[string]string, error) {
	pairs, err := ldr.Load(sources)
	if err != nil {
		return nil, errors.WrapPrefix(err, "loading KV pairs", 0)
	}
	knownKeys := make(map[string]string)
	for _, p := range pairs {
		// legal key: alphanumeric characters, '-', '_' or '.'
		if err := ldr.Validator().ErrIfInvalidKey(p.Key); err != nil {
			return nil, err
		}
		if _, ok := knownKeys[p.Key]; ok {
			return nil, errors.Errorf(
				"configmap %s illegally repeats the key `%s`", name, p.Key)
		}
		knownKeys[p.Key] = p.Value
	}
	return knownKeys, nil
}

// copyLabelsAndAnnotations copies labels and annotations from
// GeneratorOptions into the given object.
func copyLabelsAndAnnotations(
	rn *yaml.RNode, opts *types.GeneratorOptions) error {
	if opts == nil {
		return nil
	}
	for _, k := range filtersutil.SortedMapKeys(opts.Labels) {
		v := opts.Labels[k]
		if _, err := rn.Pipe(yaml.SetLabel(k, v)); err != nil {
			return err
		}
	}
	for _, k := range filtersutil.SortedMapKeys(opts.Annotations) {
		v := opts.Annotations[k]
		if _, err := rn.Pipe(yaml.SetAnnotation(k, v)); err != nil {
			return err
		}
	}
	return nil
}

// In a secret, all data is base64 encoded, regardless of its conformance
// or lack thereof to UTF-8.
func makeSecretValueRNode(s string) *yaml.RNode {
	yN := &yaml.Node{Kind: yaml.ScalarNode}
	// Purposely don't use YAML tags to identify the data as being plain text or
	// binary.  It kubernetes Secrets the values in the `data` map are expected
	// to be base64 encoded, and in ConfigMaps that same can be said for the
	// values in the `binaryData` field.
	yN.Tag = yaml.NodeTagString
	yN.Value = encodeBase64(s)
	if strings.Contains(yN.Value, "\n") {
		yN.Style = yaml.LiteralStyle
	}
	return yaml.NewRNode(yN)
}

func makeConfigMapValueRNode(s string) (field string, rN *yaml.RNode) {
	yN := &yaml.Node{Kind: yaml.ScalarNode}
	yN.Tag = yaml.NodeTagString
	if utf8.ValidString(s) {
		field = yaml.DataField
		yN.Value = s
	} else {
		field = yaml.BinaryDataField
		yN.Value = encodeBase64(s)
	}
	if strings.Contains(yN.Value, "\n") {
		yN.Style = yaml.LiteralStyle
	}
	return field, yaml.NewRNode(yN)
}

// encodeBase64 encodes s as base64 that is broken up into multiple lines
// as appropriate for the resulting length.
func encodeBase64(s string) string {
	const lineLen = 70
	encLen := base64.StdEncoding.EncodedLen(len(s))
	lines := encLen/lineLen + 1
	buf := make([]byte, encLen*2+lines)
	in := buf[0:encLen]
	out := buf[encLen:]
	base64.StdEncoding.Encode(in, []byte(s))
	k := 0
	for i := 0; i < len(in); i += lineLen {
		j := i + lineLen
		if j > len(in) {
			j = len(in)
		}
		k += copy(out[k:], in[i:j])
		if lines > 1 {
			out[k] = '\n'
			k++
		}
	}
	return string(out[:k])
}
