// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package execplugin_test

import (
	"fmt"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/internal/loadertest"
	. "sigs.k8s.io/kustomize/api/internal/plugins/execplugin"
	"sigs.k8s.io/kustomize/api/internal/plugins/loader"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
	"sigs.k8s.io/kustomize/api/types"
)

func TestExecPluginConfig(t *testing.T) {
	path := "/app"
	rf := resmap.NewFactory(
		resource.NewFactory(
			kunstruct.NewKunstructuredFactoryImpl()), nil)
	ldr := loadertest.NewFakeLoader(path)
	v := valtest_test.MakeFakeValidator()
	pluginConfig := rf.RF().FromMap(
		map[string]interface{}{
			"apiVersion": "someteam.example.com/v1",
			"kind":       "SedTransformer",
			"metadata": map[string]interface{}{
				"name": "some-random-name",
			},
			"argsOneLiner": "one two",
			"argsFromFile": "sed-input.txt",
		})

	ldr.AddFile("/app/sed-input.txt", []byte(`
s/$FOO/foo/g
s/$BAR/bar/g
 \ \ \ 
`))

	p := NewExecPlugin(
		loader.AbsolutePluginPath(
			konfig.DisabledPluginConfig(),
			pluginConfig.OrgId()))
	// Not checking to see if the plugin is executable,
	// because this test does not run it.
	// This tests only covers sending configuration
	// to the plugin wrapper object and confirming
	// that it's properly prepared for execution.

	yaml, err := pluginConfig.AsYAML()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	p.Config(resmap.NewPluginHelpers(ldr, v, rf), yaml)

	expected := "someteam.example.com/v1/sedtransformer/SedTransformer"
	if !strings.HasSuffix(p.Path(), expected) {
		t.Fatalf("expected suffix '%s', got '%s'", expected, p.Path())
	}

	expected = `apiVersion: someteam.example.com/v1
argsFromFile: sed-input.txt
argsOneLiner: one two
kind: SedTransformer
metadata:
  name: some-random-name
`
	if expected != string(p.Cfg()) {
		t.Fatalf("expected cfg '%s', got '%s'", expected, string(p.Cfg()))

	}
	if len(p.Args()) != 5 {
		t.Fatalf("unexpected arg len %d, %v", len(p.Args()), p.Args())
	}
	if p.Args()[0] != "one" ||
		p.Args()[1] != "two" ||
		p.Args()[2] != "s/$FOO/foo/g" ||
		p.Args()[3] != "s/$BAR/bar/g" ||
		p.Args()[4] != "\\ \\ \\ " {
		t.Fatalf("unexpected arg array: %v", p.Args())
	}
}

func makeConfigMap(rf *resource.Factory, name, behavior string, hashValue *string) *resource.Resource {
	r := rf.FromMap(map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata":   map[string]interface{}{"name": name},
	})
	annotations := map[string]string{}
	if behavior != "" {
		annotations[BehaviorAnnotation] = behavior
	}
	if hashValue != nil {
		annotations[HashAnnotation] = *hashValue
	}
	if len(annotations) > 0 {
		r.SetAnnotations(annotations)
	}
	return r
}

func makeConfigMapOptions(rf *resource.Factory, name, behavior string, disableHash bool) *resource.Resource {
	return rf.FromMapAndOption(map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata":   map[string]interface{}{"name": name},
	}, &types.GeneratorArgs{Behavior: behavior}, &types.GeneratorOptions{DisableNameSuffixHash: disableHash})
}

func strptr(s string) *string {
	return &s
}

func TestUpdateResourceOptions(t *testing.T) {
	p := NewExecPlugin("")
	if err := p.ErrIfNotExecutable(); err == nil {
		t.Fatalf("expected unexecutable error")
	}
	rf := resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())
	in := resmap.New()
	expected := resmap.New()
	cases := []struct {
		behavior  string
		needsHash bool
		hashValue *string
	}{
		{hashValue: strptr("false")},
		{hashValue: strptr("true"), needsHash: true},
		{behavior: "replace"},
		{behavior: "merge"},
		{behavior: "create"},
		{behavior: "nonsense"},
		{behavior: "merge", hashValue: strptr("false")},
		{behavior: "merge", hashValue: strptr("true"), needsHash: true},
	}
	for i, c := range cases {
		name := fmt.Sprintf("test%d", i)
		in.Append(makeConfigMap(rf, name, c.behavior, c.hashValue))
		expected.Append(makeConfigMapOptions(rf, name, c.behavior, !c.needsHash))
	}
	actual, err := p.UpdateResourceOptions(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err.Error())
	}
	for i, a := range expected.Resources() {
		b := actual.GetByIndex(i)
		if b == nil {
			t.Fatalf("resource %d missing from processed map", i)
		}
		if !a.Equals(b) {
			t.Errorf("expected %v got %v", a, b)
		}
		if a.NeedHashSuffix() != b.NeedHashSuffix() {
			t.Errorf("")
		}
		if a.Behavior() != b.Behavior() {
			t.Errorf("expected %v got %v", a.Behavior(), b.Behavior())
		}
	}
}

func TestUpdateResourceOptionsWithInvalidHashAnnotationValues(t *testing.T) {
	p := NewExecPlugin("")
	if err := p.ErrIfNotExecutable(); err == nil {
		t.Fatalf("expected unexecutable error")
	}
	rf := resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl())
	cases := []string{
		"",
		"FaLsE",
		"TrUe",
		"potato",
	}
	for i, c := range cases {
		name := fmt.Sprintf("test%d", i)
		in := resmap.New()
		in.Append(makeConfigMap(rf, name, "", &c))
		_, err := p.UpdateResourceOptions(in)
		if err == nil {
			t.Errorf("expected error from value %q", c)
		}
	}
}
