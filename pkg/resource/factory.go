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

package resource

import (
	"log"

	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/ifc"
	internal "sigs.k8s.io/kustomize/pkg/internal/error"
	"sigs.k8s.io/kustomize/pkg/patch"
	"sigs.k8s.io/kustomize/pkg/types"
)

// Factory makes instances of Resource.
type Factory struct {
	kf ifc.KunstructuredFactory
}

// NewFactory makes an instance of Factory.
func NewFactory(kf ifc.KunstructuredFactory) *Factory {
	return &Factory{kf: kf}
}

// FromMap returns a new instance of Resource.
func (rf *Factory) FromMap(m map[string]interface{}) *Resource {
	return &Resource{
		Kunstructured: rf.kf.FromMap(m),
		b:             ifc.BehaviorUnspecified}
}

// FromKunstructured returns a new instance of Resource.
func (rf *Factory) FromKunstructured(
	u ifc.Kunstructured) *Resource {
	if u == nil {
		log.Fatal("unstruct ifc must not be null")
	}
	return &Resource{Kunstructured: u, b: ifc.BehaviorUnspecified}
}

// SliceFromPatches returns a slice of resources given a patch path
// slice from a kustomization file.
func (rf *Factory) SliceFromPatches(
	ldr ifc.Loader, paths []patch.StrategicMerge) ([]*Resource, error) {
	var result []*Resource
	for _, path := range paths {
		content, err := ldr.Load(string(path))
		if err != nil {
			return nil, err
		}
		res, err := rf.SliceFromBytes(content)
		if err != nil {
			return nil, internal.Handler(err, string(path))
		}
		result = append(result, res...)
	}
	return result, nil
}

// SliceFromBytes unmarshalls bytes into a Resource slice.
func (rf *Factory) SliceFromBytes(in []byte) ([]*Resource, error) {
	kunStructs, err := rf.kf.SliceFromBytes(in)
	if err != nil {
		return nil, err
	}
	var result []*Resource
	for _, u := range kunStructs {
		result = append(result, rf.FromKunstructured(u))
	}
	return result, nil
}

// Set sets the filesystem and loader for the underlying factory
func (rf *Factory) Set(fs fs.FileSystem, ldr ifc.Loader) {
	rf.kf.Set(fs, ldr)
}

// MakeConfigMap makes an instance of Resource for ConfigMap
func (rf *Factory) MakeConfigMap(args *types.ConfigMapArgs) (*Resource, error) {
	u, err := rf.kf.MakeConfigMap(args)
	if err != nil {
		return nil, err
	}
	return &Resource{Kunstructured: u, b: fixBehavior(args.Behavior)}, nil
}

// MakeSecret makes an instance of Resource for Secret
func (rf *Factory) MakeSecret(args *types.SecretArgs) (*Resource, error) {
	u, err := rf.kf.MakeSecret(args)
	if err != nil {
		return nil, err
	}
	return &Resource{Kunstructured: u, b: fixBehavior(args.Behavior)}, nil
}

func fixBehavior(s string) ifc.GenerationBehavior {
	b := ifc.NewGenerationBehavior(s)
	if b == ifc.BehaviorUnspecified {
		return ifc.BehaviorCreate
	}
	return b
}
