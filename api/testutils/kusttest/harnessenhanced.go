// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kusttest_test

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"

	"sigs.k8s.io/kustomize/api/filesys"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/internal/k8sdeps/transformer"
	pLdr "sigs.k8s.io/kustomize/api/internal/plugins/loader"
	"sigs.k8s.io/kustomize/api/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/api/konfig"
	fLdr "sigs.k8s.io/kustomize/api/loader"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"
	valtest_test "sigs.k8s.io/kustomize/api/testutils/valtest"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// HarnessEnhanced manages a full plugin environment for tests.
type HarnessEnhanced struct {
	// An instance of *testing.T, and a filesystem (likely in-memory)
	// for loading test data - plugin config, resources to transform, etc.
	Harness

	// plugintestEnv holds the plugin compiler and data needed to
	// create compilation sub-processes.
	pte *pluginTestEnv

	// rf creates Resources from byte streams.
	rf *resmap.Factory

	// A file loader using the Harness.fSys to read test data.
	ldr ifc.Loader

	// A plugin loader that loads plugins from a (real) file system.
	pl *pLdr.Loader
}

func MakeEnhancedHarness(t *testing.T) *HarnessEnhanced {
	pte := newPluginTestEnv(t).set()

	pc, err := konfig.EnabledPluginConfig(types.BploLoadFromFileSys)
	if err != nil {
		t.Fatal(err)
	}

	rf := resmap.NewFactory(
		resource.NewFactory(kunstruct.NewKunstructuredFactoryImpl()),
		transformer.NewFactoryImpl())

	result := &HarnessEnhanced{
		Harness: MakeHarness(t),
		pte:     pte,
		rf:      rf,
		pl:      pLdr.NewLoader(pc, rf)}

	// Point the file loader to the root ('/') of the in-memory file system.
	result.ResetLoaderRoot(filesys.Separator)

	return result
}

func (th *HarnessEnhanced) Reset() {
	th.pte.reset()
}

func (th *HarnessEnhanced) PrepBuiltin(k string) *HarnessEnhanced {
	return th.BuildGoPlugin(konfig.BuiltinPluginPackage, "", k)
}

func (th *HarnessEnhanced) BuildGoPlugin(g, v, k string) *HarnessEnhanced {
	th.pte.prepareGoPlugin(g, v, k)
	return th
}

func (th *HarnessEnhanced) PrepExecPlugin(g, v, k string) *HarnessEnhanced {
	th.pte.prepareExecPlugin(g, v, k)
	return th
}

// ResetLoaderRoot interprets its argument as an absolute directory path.
// It creates the directory, and creates the harness's file loader
// rooted in that directory.
func (th *HarnessEnhanced) ResetLoaderRoot(root string) {
	if err := th.fSys.Mkdir(root); err != nil {
		th.t.Fatal(err)
	}
	ldr, err := fLdr.NewLoader(
		fLdr.RestrictionRootOnly, root, th.fSys)
	if err != nil {
		th.t.Fatalf("Unable to make loader: %v", err)
	}
	th.ldr = ldr
}

func (th *HarnessEnhanced) LoadAndRunGenerator(
	config string) resmap.ResMap {
	res, err := th.rf.RF().FromBytes([]byte(config))
	if err != nil {
		th.t.Fatalf("Err: %v", err)
	}
	g, err := th.pl.LoadGenerator(
		th.ldr, valtest_test.MakeFakeValidator(), res)
	if err != nil {
		th.t.Fatalf("Err: %v", err)
	}
	rm, err := g.Generate()
	if err != nil {
		th.t.Fatalf("Err: %v", err)
	}
	return rm
}

func (th *HarnessEnhanced) LoadAndRunTransformer(
	config, input string) resmap.ResMap {
	resMap, err := th.RunTransformer(config, input)
	if err != nil {
		th.t.Fatalf("Err: %v", err)
	}
	return resMap
}

func (th *HarnessEnhanced) RunTransformerAndCheckResult(
	config, input, expected string) {
	for _, b := range []bool{true, false} {
		th.t.Run(fmt.Sprintf("yaml-%v", b), func(t *testing.T) {
			c, err := toggleYamlSupportField(config, b)
			if err != nil {
				th.t.Fatalf("Err: %v", err)
			}
			resMap, err := th.RunTransformer(c, input)
			if err != nil {
				th.t.Fatalf("Err: %v", err)
			}
			th.AssertActualEqualsExpected(resMap, expected)
		})
	}
}

func toggleYamlSupportField(config string, yamlSupport bool) (string, error) {
	var out bytes.Buffer
	rw := kio.ByteReadWriter{
		Reader: bytes.NewBufferString(config),
		Writer: &out,
	}
	err := kio.Pipeline{
		Inputs: []kio.Reader{&rw},
		Filters: []kio.Filter{
			kio.FilterAll(yaml.FilterFunc(
				func(node *yaml.RNode) (*yaml.RNode, error) {
					return node.Pipe(yaml.FieldSetter{
						Name:        "yamlSupport",
						StringValue: strconv.FormatBool(yamlSupport),
					})
				}),
			),
		},
		Outputs: []kio.Writer{&rw},
	}.Execute()
	return out.String(), err
}

func (th *HarnessEnhanced) ErrorFromLoadAndRunTransformer(
	config, input string) error {
	_, err := th.RunTransformer(config, input)
	return err
}

type AssertFunc func(t *testing.T, err error)

func (th *HarnessEnhanced) RunTransformerAndCheckError(
	config, input string, assertFn AssertFunc) {
	for _, b := range []bool{true, false} {
		th.t.Run(fmt.Sprintf("yaml-%v", b), func(t *testing.T) {
			c, err := toggleYamlSupportField(config, b)
			if err != nil {
				th.t.Fatalf("Err: %v", err)
			}
			_, err = th.RunTransformer(c, input)
			assertFn(t, err)
		})
	}
}

func (th *HarnessEnhanced) RunTransformer(
	config, input string) (resmap.ResMap, error) {
	resMap, err := th.rf.NewResMapFromBytes([]byte(input))
	if err != nil {
		th.t.Fatalf("Err: %v", err)
	}
	return th.RunTransformerFromResMap(config, resMap)
}

func (th *HarnessEnhanced) RunTransformerFromResMap(
	config string, resMap resmap.ResMap) (resmap.ResMap, error) {
	transConfig, err := th.rf.RF().FromBytes([]byte(config))
	if err != nil {
		th.t.Fatalf("Err: %v", err)
	}
	g, err := th.pl.LoadTransformer(
		th.ldr, valtest_test.MakeFakeValidator(), transConfig)
	if err != nil {
		return nil, err
	}
	err = g.Transform(resMap)
	return resMap, err
}
