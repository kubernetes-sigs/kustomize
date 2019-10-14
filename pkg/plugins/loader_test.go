// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package plugins_test

import (
	"testing"

	"sigs.k8s.io/kustomize/v3/internal/loadertest"
	"sigs.k8s.io/kustomize/v3/k8sdeps/kunstruct"
	. "sigs.k8s.io/kustomize/v3/pkg/plugins"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/resource"
	"sigs.k8s.io/kustomize/v3/pkg/validators"
	"sigs.k8s.io/kustomize/v3/pluglib"
)

const (
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
	tc := pluglib.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"builtin", "", "SecretGenerator")
	tc.BuildGoPlugin(
		"someteam.example.com", "v1", "SomeServiceGenerator")

	rmF := resmap.NewFactory(resource.NewFactory(
		kunstruct.NewKunstructuredFactoryImpl()), nil)

	ldr := loadertest.NewFakeLoader("/foo")

	pLdr := NewLoader(ActivePluginConfig(), rmF)
	if pLdr == nil {
		t.Fatal("expect non-nil loader")
	}

	m, err := rmF.NewResMapFromBytes([]byte(
		someServiceGenerator + "---\n" + secretGenerator))
	if err != nil {
		t.Fatal(err)
	}

	_, err = pLdr.LoadGenerators(ldr, validators.MakeFakeValidator(), m)
	if err != nil {
		t.Fatal(err)
	}
}
