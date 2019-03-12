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

package configmapandsecret

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/kustomize/pkg/types"
)

func makeFreshSecret(
	args *types.SecretArgs) *corev1.Secret {
	s := &corev1.Secret{}
	s.APIVersion = "v1"
	s.Kind = "Secret"
	s.Name = args.Name
	s.Namespace = args.Namespace
	s.Type = corev1.SecretType(args.Type)
	if s.Type == "" {
		s.Type = corev1.SecretTypeOpaque
	}
	s.Data = map[string][]byte{}
	return s
}

// MakeSecret returns a new secret.
func (f *Factory) MakeSecret(
	args *types.SecretArgs) (*corev1.Secret, error) {
	all, err := f.loadKvPairs(args.GeneratorArgs)
	if err != nil {
		return nil, err
	}
	s := makeFreshSecret(args)
	for _, p := range all {
		err = addKvToSecret(s, p.Key, p.Value)
		if err != nil {
			return nil, err
		}
	}
	if f.options != nil {
		s.SetLabels(f.options.Labels)
		s.SetAnnotations(f.options.Annotations)
	}
	return s, nil
}

func addKvToSecret(secret *corev1.Secret, keyName, data string) error {
	if err := errIfInvalidKey(keyName); err != nil {
		return err
	}
	if _, entryExists := secret.Data[keyName]; entryExists {
		return fmt.Errorf(keyExistsErrorMsg, keyName, secret.Data)
	}
	secret.Data[keyName] = []byte(data)
	return nil
}
