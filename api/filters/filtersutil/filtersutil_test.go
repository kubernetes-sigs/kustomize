package filtersutil_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/api/filters/filtersutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestApplyToJSON(t *testing.T) {
	instance1 := bytes.NewBufferString(`{"kind": "Foo"}`)
	instance2 := bytes.NewBufferString(`{"kind": "Bar"}`)
	err := filtersutil.ApplyToJSON(
		kio.FilterFunc(func(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
			for i := range nodes {
				set := yaml.SetField(
					"foo", yaml.NewScalarRNode("bar"))
				node := nodes[i]
				err := node.PipeE(set)
				if !assert.NoError(t, err) {
					t.FailNow()
				}
			}
			return nodes, nil
		}), buffer{Buffer: instance1}, buffer{Buffer: instance2},
	)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	if !assert.Equal(t,
		strings.TrimSpace(`{"foo":"bar","kind":"Foo"}`),
		strings.TrimSpace(instance1.String())) {
		t.FailNow()
	}

	if !assert.Equal(t,
		strings.TrimSpace(`{"foo":"bar","kind":"Bar"}`),
		strings.TrimSpace(instance2.String())) {
		t.FailNow()
	}
}

type buffer struct {
	*bytes.Buffer
}

func (buff buffer) UnmarshalJSON(b []byte) error {
	buff.Reset()
	buff.Write(b)
	return nil
}

func (buff buffer) MarshalJSON() ([]byte, error) {
	return buff.Bytes(), nil
}
