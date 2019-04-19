/*
Copyright 2019 The Kubernetes Authors.

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

package prune

import (
	"fmt"
	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/k8sdeps/transformer/hash"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/transformers"
	"sigs.k8s.io/kustomize/pkg/types"
)

//const PruneAnnotation = "kustomize.k8s.io/PruneRevision"
const PruneAnnotation = "current"

// pruneTransformer compute the ConfigMap used in prune
type pruneTransformer struct {
	append      bool
	cmName      string
	cmNamespace string
}

var _ transformers.Transformer = &pruneTransformer{}

// NewPruneTransformer makes a pruneTransformer.
func NewPruneTransformer(p *types.Prune, namespace string, append bool) transformers.Transformer {
	if p == nil || p.Type != "ConfigMap" || p.ConfigMap.Namespace != namespace {
		return transformers.NewNoOpTransformer()
	}
	return &pruneTransformer{
		append:      append,
		cmName:      p.ConfigMap.Name,
		cmNamespace: p.ConfigMap.Namespace,
	}
}

// Transform generates a prune ConfigMap based on the input ResMap.
// this tranformer doesn't change existing resources -
// it just visits resources and accumulates information to make a new ConfigMap.
// The prune ConfigMap is used to support the pruning command in the client side tool,
// which is proposed in https://github.com/kubernetes/enhancements/pull/810
func (o *pruneTransformer) Transform(m resmap.ResMap) error {
	var keys []string
	for _, r := range m {
		s := r.PruneString()
		keys = append(keys, s)
		for _, refid := range r.GetRefBy() {
			ref := m[refid]
			keys = append(keys, s+"---"+ref.PruneString())
		}
	}
	h, err := hash.SortArrayAndComputeHash(keys)
	if err != nil {
		return err
	}

	args := &types.ConfigMapArgs{}
	args.Name = o.cmName
	args.Namespace = o.cmNamespace
	for _, key := range keys {
		args.LiteralSources = append(args.LiteralSources,
			key+"="+h)
	}
	opts := &types.GeneratorOptions{
		Annotations: make(map[string]string),
	}
	opts.Annotations[PruneAnnotation] = h

	kf := kunstruct.NewKunstructuredFactoryImpl()
	k, err := kf.MakeConfigMap(nil, opts, args)
	if err != nil {
		return err
	}

	if !o.append {
		for k := range m {
			delete(m, k)
		}
	}

	id := resid.NewResIdWithPrefixNamespace(
		gvk.Gvk{
			Version: "v1",
			Kind:    "ConfigMap",
		},
		o.cmName,
		"", o.cmNamespace)
	if _, ok := m[id]; ok {
		return fmt.Errorf("id %v is already used, please use a different name in the prune field", id)
	}
	m[id] = resource.NewFactory(kf).FromKunstructured(k)
	return nil
}
