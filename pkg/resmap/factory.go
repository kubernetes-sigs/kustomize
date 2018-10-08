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

package resmap

import (
	"fmt"

	"github.com/pkg/errors"
	"sigs.k8s.io/kustomize/pkg/configmapandsecret"
	"sigs.k8s.io/kustomize/pkg/ifc"
	internal "sigs.k8s.io/kustomize/pkg/internal/error"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/types"
)

// Factory makes instances of ResMap.
type Factory struct {
	resF *resource.Factory
}

// NewFactory returns a new resmap.Factory.
func NewFactory(rf *resource.Factory) *Factory {
	return &Factory{resF: rf}
}

// RF returns a resource.Factory.
func (rmF *Factory) RF() *resource.Factory {
	return rmF.resF
}

// FromFiles returns a ResMap given a resource path slice.
func (rmF *Factory) FromFiles(
	loader ifc.Loader, paths []string) (ResMap, error) {
	var result []ResMap
	for _, path := range paths {
		content, err := loader.Load(path)
		if err != nil {
			return nil, errors.Wrap(err, "Load from path "+path+" failed")
		}
		res, err := rmF.newResMapFromBytes(content)
		if err != nil {
			return nil, internal.Handler(err, path)
		}
		result = append(result, res)
	}
	return MergeWithoutOverride(result...)
}

// newResMapFromBytes decodes a list of objects in byte array format.
func (rmF *Factory) newResMapFromBytes(b []byte) (ResMap, error) {
	resources, err := rmF.resF.SliceFromBytes(b)
	if err != nil {
		return nil, err
	}

	result := ResMap{}
	for _, res := range resources {
		id := res.Id()
		if _, found := result[id]; found {
			return result, fmt.Errorf("GroupVersionKindName: %#v already exists b the map", id)
		}
		result[id] = res
	}
	return result, nil
}

// NewResMapFromConfigMapArgs returns a Resource slice given
// a configmap metadata slice from kustomization file.
func (rmF *Factory) NewResMapFromConfigMapArgs(
	cf *configmapandsecret.ConfigMapFactory,
	argList []types.ConfigMapArgs) (ResMap, error) {
	var resources []*resource.Resource
	for _, args := range argList {
		obj, err := cf.MakeConfigMap(&args)
		if err != nil {
			return nil, errors.Wrap(err, "NewResMapFromConfigMapArgs")
		}
		res, err := rmF.resF.WithBehavior(obj, fixBehavior(args.Behavior))
		if err != nil {
			return nil, err
		}
		resources = append(resources, res)
	}
	return newResMapFromResourceSlice(resources)
}

// NewResMapFromSecretArgs takes a SecretArgs slice, generates
// secrets from each entry, and accumulates them in a ResMap.
func (rmF *Factory) NewResMapFromSecretArgs(
	sf *configmapandsecret.SecretFactory,
	argsList []types.SecretArgs) (ResMap, error) {
	var resources []*resource.Resource
	for _, args := range argsList {
		obj, err := sf.MakeSecret(&args)
		if err != nil {
			return nil, errors.Wrap(err, "NewResMapFromSecretArgs")
		}
		res, err := rmF.resF.WithBehavior(obj, fixBehavior(args.Behavior))
		if err != nil {
			return nil, errors.Wrap(err, "WithBehavior")
		}
		resources = append(resources, res)
	}
	return newResMapFromResourceSlice(resources)
}

func fixBehavior(s string) ifc.GenerationBehavior {
	b := ifc.NewGenerationBehavior(s)
	if b == ifc.BehaviorUnspecified {
		return ifc.BehaviorCreate
	}
	return b
}

func newResMapFromResourceSlice(resources []*resource.Resource) (ResMap, error) {
	result := ResMap{}
	for _, res := range resources {
		id := res.Id()
		if _, found := result[id]; found {
			return nil, fmt.Errorf("duplicated %#v is not allowed", id)
		}
		result[id] = res
	}
	return result, nil
}
