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
	"sigs.k8s.io/kustomize/internal/k8sdeps/configmapandsecret"
	"sigs.k8s.io/kustomize/pkg/fs"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/types"
)

// KunstructurerFactoryImpl hides construction using apimachinery types.
type KunstructurerFactoryImpl struct {
	decoder    ifc.Decoder
	cmfactory  *configmapandsecret.ConfigMapFactory
	secfactory *configmapandsecret.SecretFactory
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

// MakeConfigMap returns an instance of Kunstructured for ConfigMap
func (kf *KunstructurerFactoryImpl) MakeConfigMap(args *types.ConfigMapArgs) (ifc.Kunstructured, error) {
	cm, err := kf.cmfactory.MakeConfigMap(args)
	if err != nil {
		return nil, err
	}
	return NewKunstructuredFromObject(cm)
}

// MakeSecret returns an instance of Kunstructured for Secret
func (kf *KunstructurerFactoryImpl) MakeSecret(args *types.SecretArgs) (ifc.Kunstructured, error) {
	sec, err := kf.secfactory.MakeSecret(args)
	if err != nil {
		return nil, err
	}
	return NewKunstructuredFromObject(sec)
}

// Set sets loader, filesystem and workdirectory
func (kf *KunstructurerFactoryImpl) Set(fs fs.FileSystem, ldr ifc.Loader) {
	kf.cmfactory = configmapandsecret.NewConfigMapFactory(fs, ldr)
	kf.secfactory = configmapandsecret.NewSecretFactory(fs, ldr.Root())
}
