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

package app

import (
	"bytes"
	"encoding/json"

	"github.com/ghodss/yaml"

	"github.com/kubernetes-sigs/kustomize/pkg/constants"
	interror "github.com/kubernetes-sigs/kustomize/pkg/internal/error"
	"github.com/kubernetes-sigs/kustomize/pkg/loader"
	"github.com/kubernetes-sigs/kustomize/pkg/resmap"
	"github.com/kubernetes-sigs/kustomize/pkg/resource"
	"github.com/kubernetes-sigs/kustomize/pkg/transformers"
	"github.com/kubernetes-sigs/kustomize/pkg/types"
)

type Application interface {
	// Resources computes and returns the resources for the app.
	Resources() (resmap.ResMap, error)
	// SemiResources computes and returns the resources without name hash and name reference for the app
	SemiResources() (resmap.ResMap, error)
	// RawResources computes and returns the raw resources from the kustomization file.
	// It contains resources from
	// 1) untransformed resources from current kustomization file
	// 2) transformed resources from sub packages
	RawResources() (resmap.ResMap, error)
}

var _ Application = &applicationImpl{}

// Private implementation of the Application interface
type applicationImpl struct {
	kustomization *types.Kustomization
	loader        loader.Loader
}

// NewApp parses the kustomization file at the path using the loader.
func New(loader loader.Loader) (Application, error) {
	content, err := loader.Load(constants.KustomizationFileName)
	if err != nil {
		return nil, err
	}

	var m types.Kustomization
	err = unmarshal(content, &m)
	if err != nil {
		return nil, err
	}
	return &applicationImpl{kustomization: &m, loader: loader}, nil
}

// Resources computes and returns the resources from the kustomization file.
// The namehashing for configmap/secrets and resolving name reference is only done
// in the most top overlay once at the end of getting resources.
func (a *applicationImpl) Resources() (resmap.ResMap, error) {
	res, err := a.SemiResources()
	if err != nil {
		return nil, err
	}
	t, err := a.getHashAndReferenceTransformer()
	if err != nil {
		return nil, err
	}
	err = t.Transform(res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// SemiResources computes and returns the resources without name hash and name reference for the app
func (a *applicationImpl) SemiResources() (resmap.ResMap, error) {
	errs := &interror.KustomizationErrors{}
	raw, err := a.rawResources()
	if err != nil {
		errs.Append(err)
	}

	cms, err := resmap.NewResMapFromConfigMapArgs(a.loader, a.kustomization.ConfigMapGenerator)
	if err != nil {
		errs.Append(err)
	}
	secrets, err := resmap.NewResMapFromSecretArgs(a.loader.Root(), a.kustomization.SecretGenerator)
	if err != nil {
		errs.Append(err)
	}
	res, err := resmap.Merge(cms, secrets)
	if err != nil {
		return nil, err
	}

	allRes, err := resmap.MergeWithOverride(raw, res)
	if err != nil {
		return nil, err
	}

	patches, err := resmap.NewResourceSliceFromPatches(a.loader, a.kustomization.Patches)
	if err != nil {
		errs.Append(err)
	}

	if len(errs.Get()) > 0 {
		return nil, errs
	}

	t, err := a.getTransformer(patches)
	if err != nil {
		return nil, err
	}
	err = t.Transform(allRes)
	if err != nil {
		return nil, err
	}

	return allRes, nil
}

// RawResources computes and returns the raw resources from the kustomization file.
// The namehashing for configmap/secrets and resolving name reference is only done
// in the most top overlay once at the end of getting resources.
func (a *applicationImpl) RawResources() (resmap.ResMap, error) {
	res, err := a.rawResources()
	if err != nil {
		return nil, err
	}
	t, err := a.getHashAndReferenceTransformer()
	if err != nil {
		return nil, err
	}
	err = t.Transform(res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (a *applicationImpl) rawResources() (resmap.ResMap, error) {
	subAppResources, errs := a.subAppResources()
	resources, err := resmap.NewResMapFromFiles(a.loader, a.kustomization.Resources)
	if err != nil {
		errs.Append(err)
	}

	if len(errs.Get()) > 0 {
		return nil, errs
	}

	return resmap.Merge(resources, subAppResources)
}

func (a *applicationImpl) subAppResources() (resmap.ResMap, *interror.KustomizationErrors) {
	sliceOfSubAppResources := []resmap.ResMap{}
	errs := &interror.KustomizationErrors{}
	for _, pkgPath := range a.kustomization.Bases {
		subloader, err := a.loader.New(pkgPath)
		if err != nil {
			errs.Append(err)
			continue
		}
		subapp, err := New(subloader)
		if err != nil {
			errs.Append(err)
			continue
		}
		// Gather all transformed resources from subpackages.
		subAppResources, err := subapp.SemiResources()
		if err != nil {
			errs.Append(err)
			continue
		}
		sliceOfSubAppResources = append(sliceOfSubAppResources, subAppResources)
	}
	allResources, err := resmap.Merge(sliceOfSubAppResources...)
	if err != nil {
		errs.Append(err)
	}
	return allResources, errs
}

// getTransformer generates the following transformers:
// 1) apply overlay
// 2) name prefix
// 3) apply labels
// 4) apply annotations
func (a *applicationImpl) getTransformer(patches []*resource.Resource) (transformers.Transformer, error) {
	ts := []transformers.Transformer{}

	ot, err := transformers.NewOverlayTransformer(patches)
	if err != nil {
		return nil, err
	}
	ts = append(ts, ot)

	ts = append(ts, transformers.NewNamespaceTransformer(string(a.kustomization.Namespace)))

	npt, err := transformers.NewDefaultingNamePrefixTransformer(string(a.kustomization.NamePrefix))
	if err != nil {
		return nil, err
	}
	ts = append(ts, npt)

	lt, err := transformers.NewDefaultingLabelsMapTransformer(a.kustomization.CommonLabels)
	if err != nil {
		return nil, err
	}
	ts = append(ts, lt)

	at, err := transformers.NewDefaultingAnnotationsMapTransformer(a.kustomization.CommonAnnotations)
	if err != nil {
		return nil, err
	}
	ts = append(ts, at)

	return transformers.NewMultiTransformer(ts), nil
}

// getHashAndReferenceTransformer generates the following transformers:
// 1) name hash for configmap and secrests
// 2) apply name reference
func (a *applicationImpl) getHashAndReferenceTransformer() (transformers.Transformer, error) {
	ts := []transformers.Transformer{}
	nht := transformers.NewNameHashTransformer()
	ts = append(ts, nht)

	nrt, err := transformers.NewDefaultingNameReferenceTransformer()
	if err != nil {
		return nil, err
	}
	ts = append(ts, nrt)
	return transformers.NewMultiTransformer(ts), nil
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
