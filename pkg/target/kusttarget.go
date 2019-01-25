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
	"log"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/pkg/constants"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/ifc/transformer"
	interror "sigs.k8s.io/kustomize/pkg/internal/error"
	patchtransformer "sigs.k8s.io/kustomize/pkg/patch/transformer"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/transformers"
	"sigs.k8s.io/kustomize/pkg/transformers/config"
	"sigs.k8s.io/kustomize/pkg/types"
)

// KustTarget encapsulates the entirety of a kustomization build.
type KustTarget struct {
	kustomization *types.Kustomization
	ldr           ifc.Loader
	fSys          fs.FileSystem
	rFactory      *resmap.Factory
	tFactory      transformer.Factory
}

// NewKustTarget returns a new instance of KustTarget primed with a Loader.
func NewKustTarget(
	ldr ifc.Loader,
	fSys fs.FileSystem,
	rFactory *resmap.Factory,
	tFactory transformer.Factory) (*KustTarget, error) {
	content, err := loadKustFile(ldr)
	if err != nil {
		return nil, err
	}
	content = types.DealWithDeprecatedFields(content)
	var k types.Kustomization
	err = unmarshal(content, &k)
	if err != nil {
		return nil, err
	}
	msgs, errs := k.EnforceFields()
	if len(errs) > 0 {
		return nil, fmt.Errorf(strings.Join(errs, "\n"))
	}
	if len(msgs) > 0 {
		log.Printf(strings.Join(msgs, "\n"))
	}
	return &KustTarget{
		kustomization: &k,
		ldr:           ldr,
		fSys:          fSys,
		rFactory:      rFactory,
		tFactory:      tFactory,
	}, nil
}

func loadKustFile(ldr ifc.Loader) ([]byte, error) {
	for _, kf := range []string{
		constants.KustomizationFileName,
		constants.SecondaryKustomizationFileName} {
		content, err := ldr.Load(kf)
		if err == nil {
			return content, nil
		}
		if !strings.Contains(err.Error(), "no such file or directory") {
			return nil, err
		}
	}
	return nil, fmt.Errorf("no kustomization.yaml file under %s", ldr.Root())
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
	ra, err := kt.accumulateTarget()
	if err != nil {
		return nil, err
	}
	if kt.shouldAddHashSuffixesToGeneratedResources() {
		// This effects only generated resources.
		// It changes only the Name field in the
		// resource held in the ResMap's value, not
		// the Name in the key in the ResMap.
		err := ra.Transform(kt.tFactory.MakeHashTransformer())
		if err != nil {
			return nil, err
		}
	}
	// Given that names have changed (prefixs/suffixes added),
	// fix all the back references to those names.
	err = ra.FixBackReferences()
	if err != nil {
		return nil, err
	}
	// With all the back references fixed, it's OK to resolve Vars.
	err = ra.ResolveVars()
	return ra.ResMap(), err
}

func (kt *KustTarget) shouldAddHashSuffixesToGeneratedResources() bool {
	return kt.kustomization.GeneratorOptions == nil ||
		!kt.kustomization.GeneratorOptions.DisableNameSuffixHash
}

// accumulateTarget returns a new ResAccumulator,
// holding customized resources and the data/rules used
// to do so.  The name back references and vars are
// not yet fixed.
func (kt *KustTarget) accumulateTarget() (
	ra *ResAccumulator, err error) {
	// TODO(monopole): Get rid of the KustomizationErrors accumulator.
	// It's not consistently used, and complicates tests.
	errs := &interror.KustomizationErrors{}
	ra, errs = kt.accumulateBases()
	resources, err := kt.rFactory.FromFiles(
		kt.ldr, kt.kustomization.Resources)
	if err != nil {
		errs.Append(errors.Wrap(err, "rawResources failed to read Resources"))
	}
	if len(errs.Get()) > 0 {
		return ra, errs
	}
	err = ra.MergeResourcesWithErrorOnIdCollision(resources)
	if err != nil {
		errs.Append(errors.Wrap(err, "MergeResourcesWithErrorOnIdCollision"))
	}
	tConfig, err := config.MakeTransformerConfig(
		kt.ldr, kt.kustomization.Configurations)
	if err != nil {
		return nil, err
	}
	err = ra.MergeConfig(tConfig)
	if err != nil {
		errs.Append(errors.Wrap(err, "MergeConfig"))
	}
	err = ra.MergeVars(kt.kustomization.Vars)
	if err != nil {
		errs.Append(errors.Wrap(err, "MergeVars"))
	}
	crdTc, err := config.LoadConfigFromCRDs(kt.ldr, kt.kustomization.Crds)
	if err != nil {
		errs.Append(errors.Wrap(err, "LoadCRDs"))
	}
	err = ra.MergeConfig(crdTc)
	if err != nil {
		errs.Append(errors.Wrap(err, "merge CRDs"))
	}
	resMap, err := kt.generateConfigMapsAndSecrets(errs)
	if err != nil {
		errs.Append(errors.Wrap(err, "generateConfigMapsAndSecrets"))
	}
	err = ra.MergeResourcesWithOverride(resMap)
	if err != nil {
		return nil, err
	}
	patches, err := kt.rFactory.RF().SliceFromPatches(
		kt.ldr, kt.kustomization.PatchesStrategicMerge)
	if err != nil {
		errs.Append(errors.Wrap(err, "SliceFromPatches"))
	}
	if len(errs.Get()) > 0 {
		return nil, errs
	}
	t, err := kt.newTransformer(patches, ra.tConfig)
	if err != nil {
		return nil, err
	}
	err = ra.Transform(t)
	if err != nil {
		return nil, err
	}
	return ra, nil
}

func (kt *KustTarget) generateConfigMapsAndSecrets(
	errs *interror.KustomizationErrors) (resmap.ResMap, error) {
	kt.rFactory.Set(kt.ldr)
	cms, err := kt.rFactory.NewResMapFromConfigMapArgs(
		kt.kustomization.ConfigMapGenerator, kt.kustomization.GeneratorOptions)
	if err != nil {
		errs.Append(errors.Wrap(err, "NewResMapFromConfigMapArgs"))
	}
	secrets, err := kt.rFactory.NewResMapFromSecretArgs(
		kt.kustomization.SecretGenerator, kt.kustomization.GeneratorOptions)
	if err != nil {
		errs.Append(errors.Wrap(err, "NewResMapFromSecretArgs"))
	}
	return resmap.MergeWithErrorOnIdCollision(cms, secrets)
}

// accumulateBases returns a new ResAccumulator
// holding customized resources and the data/rules
// used to customized them from only the _bases_
// of this KustTarget.
func (kt *KustTarget) accumulateBases() (
	ra *ResAccumulator, errs *interror.KustomizationErrors) {
	errs = &interror.KustomizationErrors{}
	ra = MakeEmptyAccumulator()

	for _, path := range kt.kustomization.Bases {
		ldr, err := kt.ldr.New(path)
		if err != nil {
			errs.Append(errors.Wrap(err, "couldn't make loader for "+path))
			continue
		}
		subKt, err := NewKustTarget(
			ldr, kt.fSys, kt.rFactory, kt.tFactory)
		if err != nil {
			errs.Append(errors.Wrap(err, "couldn't make target for "+path))
			ldr.Cleanup()
			continue
		}
		subRa, err := subKt.accumulateTarget()
		if err != nil {
			errs.Append(errors.Wrap(err, "accumulateTarget"))
			ldr.Cleanup()
			continue
		}
		err = ra.MergeAccumulator(subRa)
		if err != nil {
			errs.Append(errors.Wrap(err, path))
		}
		ldr.Cleanup()
	}
	return ra, errs
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
	t, err = transformers.NewImageTransformer(kt.kustomization.Images)
	if err != nil {
		return nil, err
	}
	r = append(r, t)
	return transformers.NewMultiTransformer(r), nil
}
