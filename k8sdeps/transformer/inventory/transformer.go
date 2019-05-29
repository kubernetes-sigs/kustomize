// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package inventory

import (
	"fmt"

	"sigs.k8s.io/kustomize/k8sdeps/kunstruct"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/hasher"
	"sigs.k8s.io/kustomize/pkg/ifc"
	"sigs.k8s.io/kustomize/pkg/inventory"
	"sigs.k8s.io/kustomize/pkg/resid"
	"sigs.k8s.io/kustomize/pkg/resmap"
	"sigs.k8s.io/kustomize/pkg/resource"
	"sigs.k8s.io/kustomize/pkg/transformers"
	"sigs.k8s.io/kustomize/pkg/types"
)

// transformer compute the inventory object used in prune
type transformer struct {
	garbagePolicy types.GarbagePolicy
	ldr           ifc.Loader
	cmName        string
	cmNamespace   string
}

var _ transformers.Transformer = &transformer{}

// NewTransformer makes a new inventory transformer.
func NewTransformer(
	p *types.Inventory,
	ldr ifc.Loader,
	namespace string,
	gp types.GarbagePolicy) transformers.Transformer {
	if p == nil || p.Type != "ConfigMap" || p.ConfigMap.Namespace != namespace {
		return transformers.NewNoOpTransformer()
	}
	return &transformer{
		garbagePolicy: gp,
		ldr:           ldr,
		cmName:        p.ConfigMap.Name,
		cmNamespace:   p.ConfigMap.Namespace,
	}
}

// Transform generates an inventory object based on the input ResMap.
// this transformer doesn't change existing resources -
// it just visits resources and accumulates information to make a new ConfigMap.
// The prune ConfigMap is used to support the pruning command in the client side tool,
// which is proposed in https://github.com/kubernetes/enhancements/pull/810
// The inventory data is written to annotation since
//   1. The key in data field is constrained and couldn't include arbitrary letters
//   2. The annotation can be put into any kind of objects
func (tf *transformer) Transform(m resmap.ResMap) error {
	invty := inventory.NewInventory()
	var keys []string
	for _, r := range m {
		ns, _ := r.GetFieldValue("metadata.namespace")
		item := resid.NewItemId(r.GetGvk(), ns, r.GetName())
		var refs []resid.ItemId

		for _, refid := range r.GetRefBy() {
			ref := m[refid]
			ns, _ := ref.GetFieldValue("metadata.namespace")
			refs = append(refs, resid.NewItemId(ref.GetGvk(), ns, ref.GetName()))
		}
		invty.Current[item] = refs
		keys = append(keys, item.String())
	}
	h, err := hasher.SortArrayAndComputeHash(keys)
	if err != nil {
		return err
	}

	args := &types.ConfigMapArgs{}
	args.Name = tf.cmName
	args.Namespace = tf.cmNamespace
	opts := &types.GeneratorOptions{
		Annotations: make(map[string]string),
	}
	opts.Annotations[inventory.HashAnnotation] = h
	err = invty.UpdateAnnotations(opts.Annotations)
	if err != nil {
		return err
	}

	kf := kunstruct.NewKunstructuredFactoryImpl()
	k, err := kf.MakeConfigMap(tf.ldr, opts, args)
	if err != nil {
		return err
	}

	if tf.garbagePolicy == types.GarbageCollect {
		for k := range m {
			delete(m, k)
		}
	}

	id := resid.NewResIdWithPrefixNamespace(
		gvk.Gvk{
			Version: "v1",
			Kind:    "ConfigMap",
		},
		tf.cmName,
		"", tf.cmNamespace)
	if _, ok := m[id]; ok {
		return fmt.Errorf("id %v is already used, please use a different name in the prune field", id)
	}
	m[id] = resource.NewFactory(kf).FromKunstructured(k)
	return nil
}
