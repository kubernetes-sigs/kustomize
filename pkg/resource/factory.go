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

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/kustomize/internal/k8sdeps"
	"sigs.k8s.io/kustomize/pkg/ifc"
	internal "sigs.k8s.io/kustomize/pkg/internal/error"
	"sigs.k8s.io/kustomize/pkg/patch"
)

// Factory makes instances of Resource.
type Factory struct {
	kf ifc.KunstructuredFactory
}

// NewFactory makes an instance of Factory.
func NewFactory(kf ifc.KunstructuredFactory) *Factory {
	return &Factory{kf: kf}
}

// WithBehavior returns a new instance of Resource.
// TODO(monopole): This runtime dependence must be refactored away.
// The logic calling this has to move to k8sdeps.
func (rf *Factory) WithBehavior(
	obj runtime.Object, b ifc.GenerationBehavior) (*Resource, error) {
	// TODO(monopole): This k8sdeps dependence must be refactored away.
	u, err := k8sdeps.NewKunstructuredFromObject(obj)
	return &Resource{Kunstructured: u, b: b}, err
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
