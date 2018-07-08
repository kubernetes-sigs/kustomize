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

package transformers

import (
	"encoding/json"
	"fmt"

	"github.com/kubernetes-sigs/kustomize/pkg/hash"
	"github.com/kubernetes-sigs/kustomize/pkg/resmap"
	"github.com/kubernetes-sigs/kustomize/pkg/resource"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// nameHashTransformer contains the prefix and the path config for each field that
// the name prefix will be applied.
type nameHashTransformer struct{
	defaultRenamingBehavior resource.RenamingBehavior
}

var _ Transformer = &nameHashTransformer{}

// NewNameHashTransformer construct a nameHashTransformer.
func NewNameHashTransformer(defaultRenamingBehavior resource.RenamingBehavior) Transformer {
	return &nameHashTransformer{ defaultRenamingBehavior: defaultRenamingBehavior }
}

// Transform appends hash to configmaps and secrets.
func (o *nameHashTransformer) Transform(m resmap.ResMap) error {
	for id, res := range m {
		if shouldNotHashName(res.RenamingBehavior(), o.defaultRenamingBehavior) {
			continue
		}

		switch {
		case selectByGVK(id.Gvk(), &schema.GroupVersionKind{Version: "v1", Kind: "ConfigMap"}):
			err := appendHashForConfigMap(res)
			if err != nil {
				return err
			}

		case selectByGVK(id.Gvk(), &schema.GroupVersionKind{Version: "v1", Kind: "Secret"}):
			err := appendHashForSecret(res)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func shouldNotHashName(resourceBehaviour resource.RenamingBehavior, defaultBehavior resource.RenamingBehavior) bool {
	return resourceBehaviour == resource.RenamingBehaviorNone ||
		(resourceBehaviour == resource.RenamingBehaviorUnspecified &&
			defaultBehavior == resource.RenamingBehaviorNone)
}

func appendHashForConfigMap(res *resource.Resource) error {
	cm, err := unstructuredToConfigmap(res)
	if err != nil {
		return err
	}

	h, err := hash.ConfigMapHash(cm)
	if err != nil {
		return err
	}
	nameWithHash := fmt.Sprintf("%s-%s", res.GetName(), h)
	res.SetName(nameWithHash)
	return nil
}

// TODO: Remove this function after we support hash unstructured objects
func unstructuredToConfigmap(res *resource.Resource) (*v1.ConfigMap, error) {
	marshaled, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}
	var out v1.ConfigMap
	err = json.Unmarshal(marshaled, &out)
	return &out, err
}

func appendHashForSecret(res *resource.Resource) error {
	secret, err := unstructuredToSecret(res)
	if err != nil {
		return err
	}

	h, err := hash.SecretHash(secret)
	if err != nil {
		return err
	}
	nameWithHash := fmt.Sprintf("%s-%s", res.GetName(), h)
	res.SetName(nameWithHash)
	return nil
}

// TODO: Remove this function after we support hash unstructured objects
func unstructuredToSecret(res *resource.Resource) (*v1.Secret, error) {
	marshaled, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}
	var out v1.Secret
	err = json.Unmarshal(marshaled, &out)
	return &out, err
}
