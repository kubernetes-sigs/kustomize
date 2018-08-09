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
	"github.com/kubernetes-sigs/kustomize/pkg/configmapandsecret"
	"github.com/kubernetes-sigs/kustomize/pkg/resource"
	"github.com/kubernetes-sigs/kustomize/pkg/types"
	"github.com/pkg/errors"
)

// NewResMapFromSecretArgs takes a SecretArgs slice, generates
// secrets from each entry, and accumulates them in a ResMap.
func NewResMapFromSecretArgs(
	f *configmapandsecret.SecretFactory,
	secretList []types.SecretArgs) (ResMap, error) {
	var allResources []*resource.Resource
	for _, args := range secretList {
		s, err := f.MakeSecret(&args)
		if err != nil {
			return nil, errors.Wrap(err, "makeSecret")
		}
		if args.Behavior == "" {
			args.Behavior = "create"
		}
		res, err := resource.NewResourceWithBehavior(
			s, resource.NewGenerationBehavior(args.Behavior))
		if err != nil {
			return nil, errors.Wrap(err, "NewResourceWithBehavior")
		}
		allResources = append(allResources, res)
	}
	return newResMapFromResourceSlice(allResources)
}
