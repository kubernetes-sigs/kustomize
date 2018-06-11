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

// Package app implements state for the set of all resources being customized.
// Should rename this - there's nothing "app"y about it.
package app

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"

	"github.com/kubernetes-sigs/kustomize/pkg/constants"
	interror "github.com/kubernetes-sigs/kustomize/pkg/internal/error"
	"github.com/kubernetes-sigs/kustomize/pkg/loader"
	"github.com/kubernetes-sigs/kustomize/pkg/resmap"
	"github.com/kubernetes-sigs/kustomize/pkg/resource"
	"github.com/kubernetes-sigs/kustomize/pkg/transformers"
	"github.com/kubernetes-sigs/kustomize/pkg/types"
	"github.com/pkg/errors"
)

// Application implements the guts of the kustomize 'build' command.
// TODO: Change name, as "application" is overloaded and somewhat
// misleading (one can customize an RBAC policy).  Perhaps "Target"
// https://github.com/kubernetes-sigs/kustomize/blob/master/docs/glossary.md#target
type Application struct {
	kustomization *types.Kustomization
	loader        loader.Loader
}

// NewApplication returns a new instance of Application primed with a Loader.
func NewApplication(loader loader.Loader) (*Application, error) {
	content, err := loader.Load(constants.KustomizationFileName)
	if err != nil {
		return nil, err
	}

	var m types.Kustomization
	err = unmarshal(content, &m)
	if err != nil {
		return nil, err
	}
	return &Application{kustomization: &m, loader: loader}, nil
}

// MakeCustomizedResMap creates a ResMap per kustomization instructions.
// The Resources in the returned ResMap are fully customized.
func (a *Application) MakeCustomizedResMap() (resmap.ResMap, error) {
	m, err := a.loadCustomizedResMap()
	if err != nil {
		return nil, err
	}
	return a.resolveRefsToGeneratedResources(m)
}

// resolveRefsToGeneratedResources fixes all name references.
func (a *Application) resolveRefsToGeneratedResources(m resmap.ResMap) (resmap.ResMap, error) {
	r := []transformers.Transformer{transformers.NewNameHashTransformer()}

	t, err := transformers.NewDefaultingNameReferenceTransformer()
	if err != nil {
		return nil, err
	}
	r = append(r, t)

	refVars, err := a.resolveRefVars(m)
	if err != nil {
		return nil, err
	}
	t, err = transformers.NewRefVarTransformer(refVars)
	if err != nil {
		return nil, err
	}
	r = append(r, t)

	err = transformers.NewMultiTransformer(r).Transform(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// loadCustomizedResMap loads and customizes resources to build a ResMap.
func (a *Application) loadCustomizedResMap() (resmap.ResMap, error) {
	errs := &interror.KustomizationErrors{}
	result, err := a.loadResMapFromBasesAndResources()
	if err != nil {
		errs.Append(errors.Wrap(err, "rawResources"))
	}

	cms, err := resmap.NewResMapFromConfigMapArgs(a.loader, a.kustomization.ConfigMapGenerator)
	if err != nil {
		errs.Append(errors.Wrap(err, "NewResMapFromConfigMapArgs"))
	}
	secrets, err := resmap.NewResMapFromSecretArgs(a.loader.Root(), a.kustomization.SecretGenerator)
	if err != nil {
		errs.Append(errors.Wrap(err, "NewResMapFromSecretArgs"))
	}
	res, err := resmap.MergeWithoutOverride(cms, secrets)
	if err != nil {
		return nil, errors.Wrap(err, "Merge")
	}

	result, err = resmap.MergeWithOverride(result, res)
	if err != nil {
		return nil, err
	}

	patches, err := resmap.NewResourceSliceFromPatches(a.loader, a.kustomization.Patches)
	if err != nil {
		errs.Append(errors.Wrap(err, "NewResourceSliceFromPatches"))
	}

	if len(errs.Get()) > 0 {
		return nil, errs
	}
	t, err := a.newTransformer(patches)
	if err != nil {
		return nil, err
	}
	err = t.Transform(result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// MakeUncustomizedResMap purports to create a ResMap without customization.
// The Resources in the returned ResMap include all resources mentioned
// in the kustomization file and transitively reachable via its Bases,
// and all generated secrets and configMaps.
// Meant for use in generating a diff against customized resources.
func (a *Application) MakeUncustomizedResMap() (resmap.ResMap, error) {
	m, err := a.loadResMapFromBasesAndResources()
	if err != nil {
		return nil, err
	}
	return a.resolveRefsToGeneratedResources(m)
}

// Gets Bases and Resources as advertised.
func (a *Application) loadResMapFromBasesAndResources() (resmap.ResMap, error) {
	bases, errs := a.loadCustomizedBases()
	resources, err := resmap.NewResMapFromFiles(a.loader, a.kustomization.Resources)
	if err != nil {
		errs.Append(errors.Wrap(err, "rawResources failed to read Resources"))
	}
	if len(errs.Get()) > 0 {
		return nil, errs
	}
	return resmap.MergeWithoutOverride(resources, bases)
}

// Loop through the Bases of this kustomization recursively loading resources.
// Combine into one ResMap, demanding unique Ids for each resource.
func (a *Application) loadCustomizedBases() (resmap.ResMap, *interror.KustomizationErrors) {
	list := []resmap.ResMap{}
	errs := &interror.KustomizationErrors{}
	for _, path := range a.kustomization.Bases {
		loader, err := a.loader.New(path)
		if err != nil {
			errs.Append(errors.Wrap(err, "couldn't make loader for "+path))
			continue
		}
		app, err := NewApplication(loader)
		if err != nil {
			errs.Append(errors.Wrap(err, "couldn't make app for "+path))
			continue
		}
		resMap, err := app.loadCustomizedResMap()
		if err != nil {
			errs.Append(errors.Wrap(err, "SemiResources"))
			continue
		}
		list = append(list, resMap)
	}
	result, err := resmap.MergeWithoutOverride(list...)
	if err != nil {
		errs.Append(errors.Wrap(err, "Merge failed"))
	}
	return result, errs
}

func (a *Application) loadBasesAsFlatList() ([]*Application, error) {
	var result []*Application
	errs := &interror.KustomizationErrors{}
	for _, path := range a.kustomization.Bases {
		loader, err := a.loader.New(path)
		if err != nil {
			errs.Append(err)
			continue
		}
		a, err := NewApplication(loader)
		if err != nil {
			errs.Append(err)
			continue
		}
		result = append(result, a)
	}
	if len(errs.Get()) > 0 {
		return nil, errs
	}
	return result, nil
}

// newTransformer makes a Transformer that does everything except resolve generated names.
func (a *Application) newTransformer(patches []*resource.Resource) (transformers.Transformer, error) {
	r := []transformers.Transformer{}
	t, err := transformers.NewPatchTransformer(patches)
	if err != nil {
		return nil, err
	}
	r = append(r, t)
	r = append(r, transformers.NewNamespaceTransformer(string(a.kustomization.Namespace)))
	t, err = transformers.NewDefaultingNamePrefixTransformer(string(a.kustomization.NamePrefix))
	if err != nil {
		return nil, err
	}
	r = append(r, t)
	t, err = transformers.NewDefaultingLabelsMapTransformer(a.kustomization.CommonLabels)
	if err != nil {
		return nil, err
	}
	r = append(r, t)
	t, err = transformers.NewDefaultingAnnotationsMapTransformer(a.kustomization.CommonAnnotations)
	if err != nil {
		return nil, err
	}
	r = append(r, t)
	return transformers.NewMultiTransformer(r), nil
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

func (a *Application) resolveRefVars(m resmap.ResMap) (map[string]string, error) {
	result := map[string]string{}
	vars, err := a.getAllVars()
	if err != nil {
		return result, err
	}
	for _, v := range vars {
		id := resource.NewResId(v.ObjRef.GroupVersionKind(), v.ObjRef.Name)
		if r, found := m[id]; found {
			s, err := r.GetFieldValue(v.FieldRef.FieldPath)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve referred var: %+v", v)
			}
			result[v.Name] = s
		} else {
			glog.Infof("couldn't resolve v: %v", v)
		}
	}
	return result, nil
}

// getAllVars returns all the "environment" style Var instances defined in the app.
func (a *Application) getAllVars() ([]types.Var, error) {
	result := []types.Var{}
	errs := &interror.KustomizationErrors{}

	bases, err := a.loadBasesAsFlatList()
	if err != nil {
		return nil, err
	}

	// TODO: computing vars and resources for bases can be combined
	for _, b := range bases {
		vars, err := b.getAllVars()
		if err != nil {
			errs.Append(err)
			continue
		}
		result = append(result, vars...)
	}
	for _, v := range a.kustomization.Vars {
		v.Defaulting()
		result = append(result, v)
	}
	if len(errs.Get()) > 0 {
		return nil, errs
	}
	return result, nil
}
