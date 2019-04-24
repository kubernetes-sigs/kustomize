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

package inventory

import (
	"fmt"
	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/k8sdeps/transformer/hash"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/inventory"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/transformers"
	"sigs.k8s.io/kustomize/pkg/types"
)

// inventoryTransformer compute the ConfigMap used in prune
type inventoryTransformer struct {
	append      bool
	cmName      string
	cmNamespace string
}

var _ transformers.Transformer = &inventoryTransformer{}

// NewInventoryTransformer makes a inventoryTransformer.
func NewInventoryTransformer(p *types.Inventory, namespace string, append bool) transformers.Transformer {
	if p == nil || p.Type != "ConfigMap" || p.ConfigMap.Namespace != namespace {
		return transformers.NewNoOpTransformer()
	}
	return &inventoryTransformer{
		append:      append,
		cmName:      p.ConfigMap.Name,
		cmNamespace: p.ConfigMap.Namespace,
	}
}

// Transform generates an inventory ConfigMap based on the input ResMap.
// this tranformer doesn't change existing resources -
// it just visits resources and accumulates information to make a new ConfigMap.
// The prune ConfigMap is used to support the pruning command in the client side tool,
// which is proposed in https://github.com/kubernetes/enhancements/pull/810
func (o *inventoryTransformer) Transform(m resmap.ResMap) error {
	invty := inventory.NewInventory()
	var keys []string
	for _, r := range m {
		ns, _ := r.GetFieldValue("metadata.namespace")
		item := resid.New(r.GetGvk(), ns, r.GetName())
		var refs []resid.ItemId

		for _, refid := range r.GetRefBy() {
			ref := m[refid]
			ns, _ := ref.GetFieldValue("metadata.namespace")
			refs = append(refs, resid.New(ref.GetGvk(), ns, ref.GetName()))
		}
		invty.Current[item.String()] = refs
		keys = append(keys, item.String())
	}
	h, err := hash.SortArrayAndComputeHash(keys)
	if err != nil {
		return err
	}

	args := &types.ConfigMapArgs{}
	args.Name = o.cmName
	args.Namespace = o.cmNamespace
	opts := &types.GeneratorOptions{
		Annotations: make(map[string]string),
	}
	opts.Annotations[inventory.InventoryHashAnnotation] = h
	err = invty.UpdateAnnotations(opts.Annotations)
	if err != nil {
		return err
	}

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
