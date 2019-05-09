/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package target implements state for the set of all resources to customize.
package target

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/pkg/accumulator"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/ifc/transformer"
	patchtransformer "sigs.k8s.io/kustomize/pkg/patch/transformer"
	"sigs.k8s.io/kustomize/pkg/pgmconfig"
	"sigs.k8s.io/kustomize/pkg/plugins"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/transformers"
	"sigs.k8s.io/kustomize/pkg/transformers/config"
	"sigs.k8s.io/kustomize/pkg/types"
	"sigs.k8s.io/yaml"
)

// KustTarget encapsulates the entirety of a kustomization build.
type KustTarget struct {
	kustomization *types.Kustomization
	ldr           ifc.Loader
	rFactory      *resmap.Factory
	tFactory      transformer.Factory
	pLdr          *plugins.Loader
}

// NewKustTarget returns a new instance of KustTarget primed with a Loader.
func NewKustTarget(
	ldr ifc.Loader,
	rFactory *resmap.Factory,
	tFactory transformer.Factory,
	pLdr *plugins.Loader) (*KustTarget, error) {
	content, err := loadKustFile(ldr)
	if err != nil {
		return nil, err
	}
	content = types.FixKustomizationPreUnmarshalling(content)
	var k types.Kustomization
	err = unmarshal(content, &k)
	if err != nil {
		return nil, err
	}
	k.FixKustomizationPostUnmarshalling()
	errs := k.EnforceFields()
	if len(errs) > 0 {
		return nil, fmt.Errorf(
			"Failed to read kustomization file under %s:\n"+
				strings.Join(errs, "\n"), ldr.Root())
	}
	return &KustTarget{
		kustomization: &k,
		ldr:           ldr,
		rFactory:      rFactory,
		tFactory:      tFactory,
		pLdr:          pLdr,
	}, nil
}

func quoted(l []string) []string {
	r := make([]string, len(l))
	for i, v := range l {
		r[i] = "'" + v + "'"
	}
	return r
}

func commaOr(q []string) string {
	return strings.Join(q[:len(q)-1], ", ") + " or " + q[len(q)-1]
}

func loadKustFile(ldr ifc.Loader) ([]byte, error) {
	var content []byte
	match := 0
	for _, kf := range pgmconfig.KustomizationFileNames {
		c, err := ldr.Load(kf)
		if err == nil {
			match += 1
			content = c
		}
	}
	switch match {
	case 0:
		return nil, fmt.Errorf(
			"unable to find one of %v in directory '%s'",
			commaOr(quoted(pgmconfig.KustomizationFileNames)), ldr.Root())
	case 1:
		return content, nil
	default:
		return nil, fmt.Errorf(
			"Found multiple kustomization files under: %s\n", ldr.Root())
	}
}

func unmarshal(y []byte, o interface{}) error {
	j, err := yaml.YAMLToJSON(y)
	if err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields()
	return dec.Decode(o)
}

// MakeCustomizedResMap creates a ResMap per kustomization instructions.
// The Resources in the returned ResMap are fully customized.
func (kt *KustTarget) MakeCustomizedResMap() (resmap.ResMap, error) {
	ra, err := kt.AccumulateTarget()
	if err != nil {
		return nil, err
	}
	err = ra.Transform(kt.tFactory.MakeHashTransformer())
	if err != nil {
		return nil, err
	}
	// Given that names have changed (prefixs/suffixes added),
	// fix all the back references to those names.
	err = ra.FixBackReferences()
	if err != nil {
		return nil, err
	}
	// With all the back references fixed, it's OK to resolve Vars.
	err = ra.ResolveVars()
	if err != nil {
		return nil, err
	}

	rm := ra.ResMap()
	pt := kt.tFactory.MakeInventoryTransformer(kt.kustomization.Inventory, kt.kustomization.Namespace, true)
	err = pt.Transform(rm)
	if err != nil {
		return nil, err
	}
	return rm, nil
}

func (kt *KustTarget) MakePruneConfigMap() (resmap.ResMap, error) {
	ra, err := kt.AccumulateTarget()
	if err != nil {
		return nil, err
	}
	err = ra.Transform(kt.tFactory.MakeHashTransformer())
	if err != nil {
		return nil, err
	}
	// Given that names have changed (prefixs/suffixes added),
	// fix all the back references to those names.
	err = ra.FixBackReferences()
	if err != nil {
		return nil, err
	}
	// With all the back references fixed, it's OK to resolve Vars.
	err = ra.ResolveVars()
	if err != nil {
		return nil, err
	}

	rm := ra.ResMap()
	pt := kt.tFactory.MakeInventoryTransformer(kt.kustomization.Inventory, kt.kustomization.Namespace, false)
	err = pt.Transform(rm)
	if err != nil {
		return nil, err
	}
	return rm, nil
}

func (kt *KustTarget) shouldAddHashSuffixesToGeneratedResources() bool {
	return kt.kustomization.GeneratorOptions == nil ||
		!kt.kustomization.GeneratorOptions.DisableNameSuffixHash
}

// AccumulateTarget returns a new ResAccumulator,
// holding customized resources and the data/rules used
// to do so.  The name back references and vars are
// not yet fixed.
func (kt *KustTarget) AccumulateTarget() (
	ra *accumulator.ResAccumulator, err error) {
	ra = accumulator.MakeEmptyAccumulator()
	err = kt.accumulateResources(ra, kt.kustomization.Bases)
	if err != nil {
		return nil, errors.Wrap(err, "accumulating bases")
	}
	err = kt.accumulateResources(ra, kt.kustomization.Resources)
	if err != nil {
		return nil, errors.Wrap(err, "accumulating resources")
	}
	tConfig, err := config.MakeTransformerConfig(
		kt.ldr, kt.kustomization.Configurations)
	if err != nil {
		return nil, err
	}
	err = ra.MergeConfig(tConfig)
	if err != nil {
		return nil, errors.Wrapf(
			err, "merging config %v", tConfig)
	}
	err = ra.MergeVars(kt.kustomization.Vars)
	if err != nil {
		return nil, errors.Wrapf(
			err, "merging vars %v", kt.kustomization.Vars)
	}
	crdTc, err := config.LoadConfigFromCRDs(kt.ldr, kt.kustomization.Crds)
	if err != nil {
		return nil, errors.Wrapf(
			err, "loading CRDs %v", kt.kustomization.Crds)
	}
	err = ra.MergeConfig(crdTc)
	if err != nil {
		return nil, errors.Wrapf(
			err, "merging CRDs %v", crdTc)
	}
	resMap, err := kt.generateConfigMapsAndSecrets()
	if err != nil {
		return nil, errors.Wrap(
			err, "generating legacy configMaps and secrets")
	}
	err = ra.MergeResourcesWithOverride(resMap)
	if err != nil {
		return nil, errors.Wrap(
			err, "merging legacy configMaps and secrets")
	}
	err = kt.generateFromPlugins(ra)
	if err != nil {
		return nil, err
	}
	patches, err := kt.rFactory.RF().SliceFromPatches(
		kt.ldr, kt.kustomization.PatchesStrategicMerge)
	if err != nil {
		return nil, errors.Wrapf(
			err, "reading strategic merge patches %v",
			kt.kustomization.PatchesStrategicMerge)
	}
	t, err := kt.newTransformer(patches, ra.GetTransformerConfig())
	if err != nil {
		return nil, err
	}
	err = ra.Transform(t)
	if err != nil {
		return nil, err
	}
	return ra, nil
}

func (kt *KustTarget) generateFromPlugins(
	ra *accumulator.ResAccumulator) error {
	generators, err := kt.loadGeneratorPlugins()
	if err != nil {
		return errors.Wrap(err, "loading generator plugins")
	}
	for _, g := range generators {
		resMap, err := g.Generate()
		if err != nil {
			return errors.Wrapf(err, "generating from %v", g)
		}
		err = ra.MergeResourcesWithErrorOnIdCollision(resMap)
		if err != nil {
			return errors.Wrapf(err, "merging from generator %v", g)
		}
	}
	return nil
}

func (kt *KustTarget) generateConfigMapsAndSecrets() (resmap.ResMap, error) {
	cms, err := kt.rFactory.NewResMapFromConfigMapArgs(
		kt.ldr,
		kt.kustomization.GeneratorOptions,
		kt.kustomization.ConfigMapGenerator)
	if err != nil {
		return nil, errors.Wrapf(
			err, "configmapgenerator: %v", kt.kustomization.ConfigMapGenerator)
	}
	secrets, err := kt.rFactory.NewResMapFromSecretArgs(
		kt.ldr,
		kt.kustomization.GeneratorOptions,
		kt.kustomization.SecretGenerator)
	if err != nil {
		return nil, errors.Wrapf(
			err, "secretgenerator: %v", kt.kustomization.SecretGenerator)
	}
	return resmap.MergeWithErrorOnIdCollision(cms, secrets)
}

// accumulateResources fills the given resourceAccumulator
// with resources read from the given list of paths.
func (kt *KustTarget) accumulateResources(
	ra *accumulator.ResAccumulator, paths []string) error {
	for _, path := range paths {
		ldr, err := kt.ldr.New(path)
		if err == nil {
			err = kt.accumulateDirectory(ra, ldr, path)
			if err != nil {
				return err
			}
		} else {
			err = kt.accumulateFile(ra, path)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (kt *KustTarget) accumulateDirectory(
	ra *accumulator.ResAccumulator, ldr ifc.Loader, path string) error {
	defer ldr.Cleanup()
	subKt, err := NewKustTarget(
		ldr, kt.rFactory, kt.tFactory, kt.pLdr)
	if err != nil {
		return errors.Wrapf(err, "couldn't make target for path '%s'", path)
	}
	subRa, err := subKt.AccumulateTarget()
	if err != nil {
		return errors.Wrapf(
			err, "recursed accumulation of path '%s'", path)
	}
	err = ra.MergeAccumulator(subRa)
	if err != nil {
		return errors.Wrapf(
			err, "recursed merging from path '%s'", path)
	}
	return nil
}

func (kt *KustTarget) accumulateFile(
	ra *accumulator.ResAccumulator, path string) error {
	resources, err := kt.rFactory.FromFile(kt.ldr, path)
	if err != nil {
		return errors.Wrapf(err, "accumulating resources from '%s'", path)
	}
	err = ra.MergeResourcesWithErrorOnIdCollision(resources)
	if err != nil {
		return errors.Wrapf(err, "merging resources from '%s'", path)
	}
	return nil
}

// newTransformer makes a Transformer that does a collection
// of object transformations.
func (kt *KustTarget) newTransformer(
	patches []*resource.Resource, tConfig *config.TransformerConfig) (
	transformers.Transformer, error) {
	var r []transformers.Transformer
	t, err := kt.tFactory.MakePatchTransformer(patches, kt.rFactory.RF())
	if err != nil {
		return nil, err
	}
	r = append(r, t)
	r = append(r, transformers.NewNamespaceTransformer(
		string(kt.kustomization.Namespace), tConfig.NameSpace))
	t, err = transformers.NewNamePrefixSuffixTransformer(
		string(kt.kustomization.NamePrefix),
		string(kt.kustomization.NameSuffix),
		tConfig.NamePrefix,
	)
	if err != nil {
		return nil, err
	}
	r = append(r, t)
	t, err = transformers.NewLabelsMapTransformer(
		kt.kustomization.CommonLabels, tConfig.CommonLabels)
	if err != nil {
		return nil, err
	}
	r = append(r, t)
	t, err = transformers.NewAnnotationsMapTransformer(
		kt.kustomization.CommonAnnotations, tConfig.CommonAnnotations)
	if err != nil {
		return nil, err
	}
	r = append(r, t)
	t, err = patchtransformer.NewPatchJson6902Factory(kt.ldr).
		MakePatchJson6902Transformer(kt.kustomization.PatchesJson6902)
	if err != nil {
		return nil, err
	}
	r = append(r, t)
	t, err = transformers.NewImageTransformer(kt.kustomization.Images, tConfig.Images)
	if err != nil {
		return nil, err
	}
	r = append(r, t)
	tp, err := kt.loadTransformerPlugins()
	if err != nil {
		return nil, err
	}
	r = append(r, tp...)
	return transformers.NewMultiTransformer(r), nil
}

func (kt *KustTarget) loadTransformerPlugins() ([]transformers.Transformer, error) {
	ra := accumulator.MakeEmptyAccumulator()
	err := kt.accumulateResources(ra, kt.kustomization.Transformers)
	if err != nil {
		return nil, err
	}
	return kt.pLdr.LoadAndCOnfigureTransformers(kt.ldr, ra.ResMap())
}

func (kt *KustTarget) loadGeneratorPlugins() ([]transformers.Generator, error) {
	ra := accumulator.MakeEmptyAccumulator()
	err := kt.accumulateResources(ra, kt.kustomization.Generators)
	if err != nil {
		return nil, err
	}
	return kt.pLdr.LoadAndConfigureGenerators(kt.ldr, ra.ResMap())
}
