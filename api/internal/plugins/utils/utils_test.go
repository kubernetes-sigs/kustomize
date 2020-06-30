// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	"sigs.k8s.io/kustomize/api/types"
)

func TestDeterminePluginSrcRoot(t *testing.T) {
	actual, err := DeterminePluginSrcRoot(filesys.MakeFsOnDisk())
	if err != nil {
		t.Error(err)
	}
	if !filepath.IsAbs(actual) {
		t.Errorf("expected absolute path, but got '%s'", actual)
	}
	if !strings.HasSuffix(actual, konfig.RelPluginHome) {
		t.Errorf("expected suffix '%s' in '%s'", konfig.RelPluginHome, actual)
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
	}, &types.GeneratorArgs{
		Behavior: behavior,
		Options:  &types.GeneratorOptions{DisableNameSuffixHash: disableHash}})
}

func strptr(s string) *string {
	return &s
}

func TestUpdateResourceOptions(t *testing.T) {
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
	actual, err := UpdateResourceOptions(in)
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
		_, err := UpdateResourceOptions(in)
		if err == nil {
			t.Errorf("expected error from value %q", c)
		}
	}
}
