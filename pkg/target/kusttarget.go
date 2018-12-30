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
	"sigs.k8s.io/kustomize/pkg/resid"
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

// CustomizedResMap holds a transformed ResMap and the
// configuration rules used to transform it.
// At the time of writing, the ResMap held does not have
// its final name back references fixed and vars replaced.
type CustomizedResMap struct {
	resMap  resmap.ResMap
	tConfig *config.TransformerConfig
	varMap  map[string]types.Var
}

// NewKustTarget returns a new instance of KustTarget primed with a Loader.
func NewKustTarget(
	ldr ifc.Loader, fSys fs.FileSystem,
	rFactory *resmap.Factory,
	tFactory transformer.Factory) (*KustTarget, error) {
	content, err := loadKustFile(ldr)
	if err != nil {
		return nil, err
	}

	var k types.Kustomization
	err = unmarshal(content, &k)
	if err != nil {
		return nil, err
	}
	k.DealWithDeprecatedFields()
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

func unmarshal(y []byte, o interface{}) error {
	j, err := yaml.YAMLToJSON(y)
	if err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(j))
	dec.DisallowUnknownFields()
	return dec.Decode(o)
}

// TODO(#6060) Maybe switch to the false path permanently
// (desired by #606), or expose this as a new customization
// directive.
const demandExplicitConfig = true

func makeTransformerConfig(
	ldr ifc.Loader, paths []string) (*config.TransformerConfig, error) {
	if demandExplicitConfig {
		return loadConfigFromDiskOrDefaults(ldr, paths)
	}
	return mergeCustomConfigWithDefaults(ldr, paths)
}

// loadConfigFromDiskOrDefaults returns a TransformerConfig object
// built from either files or the hardcoded default configs.
// There's no merging, it's one or the other.  This is preferred if one
// wants all configuration to be explicit in version control, as
// opposed to relying on a mix of files and hard coded config.
func loadConfigFromDiskOrDefaults(
	ldr ifc.Loader, paths []string) (*config.TransformerConfig, error) {
	if paths == nil || len(paths) == 0 {
		return config.NewFactory(nil).DefaultConfig(), nil
	}
	return config.NewFactory(ldr).FromFiles(paths)
}

// mergeCustomConfigWithDefaults returns a merger of custom config,
// if any, with default config.
func mergeCustomConfigWithDefaults(
	ldr ifc.Loader, paths []string) (*config.TransformerConfig, error) {
	t1 := config.NewFactory(nil).DefaultConfig()
	if len(paths) == 0 {
		return t1, nil
	}
	t2, err := config.NewFactory(ldr).FromFiles(paths)
	if err != nil {
		return nil, err
	}
	return t1.Merge(t2)
}

// MakeCustomizedResMap creates a ResMap per kustomization instructions.
// The Resources in the returned ResMap are fully customized.
func (kt *KustTarget) MakeCustomizedResMap() (resmap.ResMap, error) {
	cr, err := kt.loadCustomizedResMap()
	if err != nil {
		return nil, err
	}
	if kt.shouldAddHashSuffixesToGeneratedResources() {
		// This effects only generated resources.
		// It changes only the Name field in the
		// resource held in the ResMap's value, not
		// the Name in the key in the ResMap.
		err := kt.tFactory.MakeHashTransformer().Transform(cr.resMap)
		if err != nil {
			return nil, err
		}
	}
	// Given that names have changed (prefixs/suffixes added),
	// fix all the back references to those names.
	if cr.tConfig.NameReference != nil {
		err = transformers.NewNameReferenceTransformer(
			cr.tConfig.NameReference).Transform(cr.resMap)
		if err != nil {
			return nil, err
		}
	}
	// With all the back references fixed, it's OK to resolve Vars.
	err = kt.doVarReplacement(cr)
	return cr.resMap, err
}

func (kt *KustTarget) shouldAddHashSuffixesToGeneratedResources() bool {
	return kt.kustomization.GeneratorOptions == nil ||
		!kt.kustomization.GeneratorOptions.DisableNameSuffixHash
}

func (kt *KustTarget) doVarReplacement(cr *CustomizedResMap) error {
	varMap, err := kt.resolveVars(cr.resMap, cr.varMap)
	if err != nil {
		return err
	}
	return transformers.NewRefVarTransformer(
		varMap, cr.tConfig.VarReference).Transform(cr.resMap)
}

// loadCustomizedResMap returns a new CustomizedResMap,
// holding customized resources and the data/rules used
// to do so.  The name back references and vars are
// not yet fixed.
func (kt *KustTarget) loadCustomizedResMap() (
	cr *CustomizedResMap, err error) {
	errs := &interror.KustomizationErrors{}

	cr, err = kt.loadResMapFromBasesAndResources()
	if err != nil {
		errs.Append(errors.Wrap(err, "loadResMapFromBasesAndResources"))
		if cr == nil {
			return nil, errs
		}
	}
	crdTc, err := config.NewFactory(kt.ldr).LoadCRDs(kt.kustomization.Crds)
	if err != nil {
		errs.Append(errors.Wrap(err, "LoadCRDs"))
	}
	cr.tConfig, err = cr.tConfig.Merge(crdTc)
	if err != nil {
		errs.Append(errors.Wrap(err, "merge CRDs"))
	}
	resMap, err := kt.generateConfigMapsAndSecrets(errs)
	if err != nil {
		errs.Append(errors.Wrap(err, "generateConfigMapsAndSecrets"))
	}
	cr.resMap, err = resmap.MergeWithOverride(cr.resMap, resMap)
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

	var r []transformers.Transformer
	t, err := kt.newTransformer(patches, cr.tConfig)
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
	t, err = transformers.NewImageTagTransformer(kt.kustomization.ImageTags)
	if err != nil {
		return nil, err
	}
	r = append(r, t)

	err = transformers.NewMultiTransformer(r).Transform(cr.resMap)
	if err != nil {
		return nil, err
	}
	return cr, nil
}

func (kt *KustTarget) generateConfigMapsAndSecrets(
	errs *interror.KustomizationErrors) (resmap.ResMap, error) {
	kt.rFactory.Set(kt.fSys, kt.ldr)
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

// loadResMapFromBasesAndResources loads this KustTarget's
// bases, customizing them per their rules, and loads this
// KustTarget's resources and config data, not yet doing
// anything with them. The uncustomized resources from this
// level and the customized resources from the bases are
// combined into one ResMap to be processed later per the
// configuration loaded at this level.
func (kt *KustTarget) loadResMapFromBasesAndResources() (
	cr *CustomizedResMap, err error) {
	cr, errs := kt.loadCustomizedBases()

	// Merge resources.
	resources, err := kt.rFactory.FromFiles(
		kt.ldr, kt.kustomization.Resources)
	if err != nil {
		errs.Append(errors.Wrap(err, "rawResources failed to read Resources"))
	}
	if len(errs.Get()) > 0 {
		return cr, errs
	}
	cr.resMap, err = resmap.MergeWithErrorOnIdCollision(
		resources, cr.resMap)
	if err != nil {
		return cr, err
	}

	// Merge config.
	tConfig, err := makeTransformerConfig(
		kt.ldr, kt.kustomization.Configurations)
	if err != nil {
		return cr, err
	}
	cr.tConfig, err = cr.tConfig.Merge(tConfig)

	// Merge vars.
	for _, v := range kt.kustomization.Vars {
		v.Defaulting()
		if _, oops := cr.varMap[v.Name]; oops {
			return nil, ErrVarCollision{v.Name, kt.ldr.Root()}
		}
		cr.varMap[v.Name] = v
	}
	return cr, err
}

// loadCustomizedBases returns a new CustomizedResMap
// holding customized resources and the data/rules
// used to customized them from only the _bases_
// of this KustTarget.
func (kt *KustTarget) loadCustomizedBases() (
	cr *CustomizedResMap, errs *interror.KustomizationErrors) {
	errs = &interror.KustomizationErrors{}
	cr = &CustomizedResMap{}
	cr.resMap = make(resmap.ResMap)
	cr.tConfig = &config.TransformerConfig{}
	cr.varMap = make(map[string]types.Var)

	for _, path := range kt.kustomization.Bases {
		ldr, err := kt.ldr.New(path)
		if err != nil {
			errs.Append(errors.Wrap(err, "couldn't make loader for "+path))
			continue
		}
		target, err := NewKustTarget(
			ldr, kt.fSys, kt.rFactory, kt.tFactory)
		if err != nil {
			errs.Append(errors.Wrap(err, "couldn't make target for "+path))
			continue
		}
		subCr, err := target.loadCustomizedResMap()
		if err != nil {
			errs.Append(errors.Wrap(err, "loadCustomizedResMap"))
			continue
		}
		ldr.Cleanup()
		// Merge resources.
		cr.resMap, err = resmap.MergeWithErrorOnIdCollision(
			cr.resMap, subCr.resMap)
		if err != nil {
			errs.Append(errors.Wrap(err, "resource merge failed"))
		}
		// Merge config.
		cr.tConfig, err = cr.tConfig.Merge(subCr.tConfig)
		if err != nil {
			errs.Append(errors.Wrap(err, "config merge failed"))
		}
		// Merge vars.
		for k, v := range subCr.varMap {
			if _, oops := cr.varMap[k]; oops {
				errs.Append(ErrVarCollision{v.Name, target.ldr.Root()})
				continue
			}
			cr.varMap[k] = v
		}
	}
	return cr, errs
}

func (kt *KustTarget) loadBasesAsKustTargets() ([]*KustTarget, error) {
	var result []*KustTarget
	for _, path := range kt.kustomization.Bases {
		ldr, err := kt.ldr.New(path)
		if err != nil {
			return nil, err
		}
		target, err := NewKustTarget(
			ldr, kt.fSys, kt.rFactory, kt.tFactory)
		if err != nil {
			return nil, err
		}
		result = append(result, target)
	}
	return result, nil
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
	return transformers.NewMultiTransformer(r), nil
}

// resolveVars returns a map of Var names to their final values.
// The values are strings intended for substitution wherever
// the $(var.Name) occurs.
func (kt *KustTarget) resolveVars(
	m resmap.ResMap, vars map[string]types.Var) (map[string]string, error) {
	result := map[string]string{}
	for _, v := range vars {
		id := resid.NewResId(v.ObjRef.GVK(), v.ObjRef.Name)
		if r, found := m.DemandOneMatchForId(id); found {
			s, err := r.GetFieldValue(v.FieldRef.FieldPath)
			if err != nil {
				return nil, fmt.Errorf("field path err for var: %+v", v)
			}
			result[v.Name] = s
		} else {
			log.Printf("couldn't resolve var: %v", v)
		}
	}
	return result, nil
}

type ErrVarCollision struct {
	name string
	path string
}

func (e ErrVarCollision) Error() string {
	return fmt.Sprintf(
		"var %s in %s defined in some other kustomization",
		e.name, e.path)
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
