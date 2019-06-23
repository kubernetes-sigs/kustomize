// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package kunstruct

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/kustomize/v3/k8sdeps/configmapandsecret"
	"sigs.k8s.io/kustomize/v3/pkg/ifc"
	"sigs.k8s.io/kustomize/v3/pkg/types"
)

// KunstructuredFactoryImpl hides construction using apimachinery types.
type KunstructuredFactoryImpl struct {
	hasher *kustHash
}

var _ ifc.KunstructuredFactory = &KunstructuredFactoryImpl{}

// NewKunstructuredFactoryImpl returns a factory.
func NewKunstructuredFactoryImpl() ifc.KunstructuredFactory {
	return &KunstructuredFactoryImpl{hasher: NewKustHash()}
}

// Hasher returns a kunstructured hasher
// input: kunstructured; output: string hash.
func (kf *KunstructuredFactoryImpl) Hasher() ifc.KunstructuredHasher {
	return kf.hasher
}

// SliceFromBytes returns a slice of Kunstructured.
func (kf *KunstructuredFactoryImpl) SliceFromBytes(
	in []byte) ([]ifc.Kunstructured, error) {
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(in), 1024)
	var result []ifc.Kunstructured
	var err error
	for err == nil || isEmptyYamlError(err) {
		var out unstructured.Unstructured
		err = decoder.Decode(&out)
		if err == nil {
			if len(out.Object) == 0 {
				continue
			}
			err = kf.validate(out)
			if err != nil {
				return nil, err
			}
			result = append(result, &UnstructAdapter{Unstructured: out})
		}
	}
	if err != io.EOF {
		return nil, err
	}
	return result, nil
}

func isEmptyYamlError(err error) bool {
	return strings.Contains(err.Error(), "is missing in 'null'")
}

// FromMap returns an instance of Kunstructured.
func (kf *KunstructuredFactoryImpl) FromMap(
	m map[string]interface{}) ifc.Kunstructured {
	return &UnstructAdapter{Unstructured: unstructured.Unstructured{Object: m}}
}

// MakeConfigMap returns an instance of Kunstructured for ConfigMap
func (kf *KunstructuredFactoryImpl) MakeConfigMap(
	ldr ifc.Loader,
	options *types.GeneratorOptions,
	args *types.ConfigMapArgs) (ifc.Kunstructured, error) {
	o, err := configmapandsecret.NewFactory(
		ldr, options).MakeConfigMap(args)
	if err != nil {
		return nil, err
	}
	return NewKunstructuredFromObject(o)
}

// MakeSecret returns an instance of Kunstructured for Secret
func (kf *KunstructuredFactoryImpl) MakeSecret(
	ldr ifc.Loader,
	options *types.GeneratorOptions,
	args *types.SecretArgs) (ifc.Kunstructured, error) {
	o, err := configmapandsecret.NewFactory(
		ldr, options).MakeSecret(args)
	if err != nil {
		return nil, err
	}
	return NewKunstructuredFromObject(o)
}

// validate validates that u has kind and name
// except for kind `List`, which doesn't require a name
func (kf *KunstructuredFactoryImpl) validate(u unstructured.Unstructured) error {
	kind := u.GetKind()
	if kind == "" {
		return fmt.Errorf("missing kind in object %v", u)
	} else if strings.HasSuffix(kind, "List") {
		return nil
	}
	if u.GetName() == "" {
		return fmt.Errorf("missing metadata.name in object %v", u)
	}
	return nil
}
