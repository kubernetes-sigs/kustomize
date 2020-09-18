// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package loader_test

import (
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	. "sigs.k8s.io/kustomize/api/internal/plugins/loader"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
	"sigs.k8s.io/kustomize/api/types"
)

const (
	//nolint:gosec
	secretGenerator = `
apiVersion: builtin
kind: SecretGenerator
metadata:
  name: secretGenerator
name: mySecret
behavior: merge
envFiles:
- a.env
- b.env
valueFiles:
- longsecret.txt
literals:
- FRUIT=apple
- VEGETABLE=carrot
`
	someServiceGenerator = `
apiVersion: someteam.example.com/v1
kind: SomeServiceGenerator
metadata:
  name: myServiceGenerator
service: my-service
port: "12345"
`
)

func TestLoader(t *testing.T) {
	th := kusttest_test.MakeEnhancedHarness(t).
		BuildGoPlugin("builtin", "", "SecretGenerator").
		BuildGoPlugin("someteam.example.com", "v1", "SomeServiceGenerator")
	defer th.Reset()
	rmF := resmap.NewFactory(resource.NewFactory(
		kunstruct.NewKunstructuredFactoryImpl()), nil)
	fLdr, err := loader.NewLoader(
		loader.RestrictionRootOnly,
		filesys.Separator, filesys.MakeFsInMemory())
	if err != nil {
		t.Fatal(err)
	}
	generatorConfigs, err := rmF.NewResMapFromBytes([]byte(
		someServiceGenerator + "---\n" + secretGenerator))
	if err != nil {
		t.Fatal(err)
	}
	for _, behavior := range []types.BuiltinPluginLoadingOptions{
		/* types.BploUseStaticallyLinked,
		types.BploLoadFromFileSys */} {
		c, err := konfig.EnabledPluginConfig(behavior)
		if err != nil {
			t.Fatal(err)
		}
		pLdr := NewLoader(c, rmF)
		if pLdr == nil {
			t.Fatal("expect non-nil loader")
		}
		_, err = pLdr.LoadGenerators(
			fLdr, valtest_test.MakeFakeValidator(), generatorConfigs)
		if err != nil {
			t.Fatal(err)
		}
	}
}
