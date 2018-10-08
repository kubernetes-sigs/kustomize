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

package k8sdeps

import (
	"io"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/kustomize/pkg/ifc"
)

// KunstructurerFactoryImpl hides construction using apimachinery types.
type KunstructurerFactoryImpl struct {
	decoder ifc.Decoder
}

var _ ifc.KunstructuredFactory = &KunstructurerFactoryImpl{}

// NewKunstructuredFactoryImpl returns a factory.
func NewKunstructuredFactoryImpl(d ifc.Decoder) ifc.KunstructuredFactory {
	return &KunstructurerFactoryImpl{decoder: d}
}

// SliceFromBytes returns a slice of Kunstructured.
func (kf *KunstructurerFactoryImpl) SliceFromBytes(
	in []byte) ([]ifc.Kunstructured, error) {
	kf.decoder.SetInput(in)
	var result []ifc.Kunstructured
	var err error
	for err == nil || isEmptyYamlError(err) {
		var out unstructured.Unstructured
		err = kf.decoder.Decode(&out)
		if err == nil {
			result = append(result, &UnstructAdapter{Unstructured: out})
		}
	}
	if err != io.EOF {
		return nil, err
	}
	return result, nil
}

// FromMap returns an instance of Kunstructured.
func (kf *KunstructurerFactoryImpl) FromMap(
	m map[string]interface{}) ifc.Kunstructured {
	return &UnstructAdapter{Unstructured: unstructured.Unstructured{Object: m}}
}
