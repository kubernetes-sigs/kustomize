// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package wrappy

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/go-errors/errors"
	"sigs.k8s.io/kustomize/api/ifc"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// WNodeFactory makes instances of WNode.
//
// These instances in turn adapt
//   sigs.k8s.io/kustomize/kyaml/yaml.RNode
// to implement ifc.Unstructured.
// This factory is meant to implement ifc.KunstructuredFactory.
//
// This implementation should be thin, as both WNode and WNodeFactory must be
// factored away (deleted) along with ifc.Kunstructured in favor of direct use
// of RNode methods upon completion of
// https://github.com/kubernetes-sigs/kustomize/issues/2506.
//
// See also api/krusty/internal/provider/depprovider.go
type WNodeFactory struct {
}

var _ ifc.KunstructuredFactory = (*WNodeFactory)(nil)

func (k *WNodeFactory) SliceFromBytes(bs []byte) ([]ifc.Kunstructured, error) {
	r := kio.ByteReader{OmitReaderAnnotations: true}
	r.Reader = bytes.NewBuffer(bs)
	yamlRNodes, err := r.Read()
	if err != nil {
		return nil, err
	}
	var result []ifc.Kunstructured
	for i := range yamlRNodes {
		rn := yamlRNodes[i]
		meta, err := rn.GetValidatedMetadata()
		if err != nil {
			return nil, err
		}
		if !shouldDropObject(meta) {
			if foundNil, path := rn.HasNilEntryInList(); foundNil {
				return nil, fmt.Errorf("empty item at %v in object %v", path, rn)
			}
			result = append(result, FromRNode(rn))
		}
	}
	return result, nil
}

// shouldDropObject returns true if the resource should not be accumulated.
func shouldDropObject(m yaml.ResourceMeta) bool {
	_, y := m.ObjectMeta.Annotations[konfig.IgnoredByKustomizeResourceAnnotation]
	return y
}

func (k *WNodeFactory) FromMap(m map[string]interface{}) ifc.Kunstructured {
	rn, err := FromMap(m)
	if err != nil {
		// TODO(#WNodeFactory): handle or bubble error"
		panic(err)
	}
	return rn
}

func (k *WNodeFactory) Hasher() ifc.KunstructuredHasher {
	panic("TODO(#WNodeFactory): implement Hasher")
}

func (k *WNodeFactory) MakeConfigMap(
	ldr ifc.KvLoader, args *types.ConfigMapArgs) (ifc.Kunstructured, error) {
	rn, err := k.makeConfigMap(ldr, args)
	if err != nil {
		return nil, err
	}
	return FromRNode(rn), nil
}

func (k *WNodeFactory) makeConfigMap(
	ldr ifc.KvLoader, args *types.ConfigMapArgs) (*yaml.RNode, error) {
	rn, err := yaml.Parse(`
apiVersion: v1
kind: ConfigMap
`)
	if err != nil {
		return nil, err
	}
	err = applyGeneratorArgs(rn, ldr, args.GeneratorArgs)
	return rn, err
}

func (k *WNodeFactory) MakeSecret(
	ldr ifc.KvLoader, args *types.SecretArgs) (ifc.Kunstructured, error) {
	rn, err := k.makeSecret(ldr, args)
	if err != nil {
		return nil, err
	}
	return FromRNode(rn), nil
}

func (k *WNodeFactory) makeSecret(
	ldr ifc.KvLoader, args *types.SecretArgs) (*yaml.RNode, error) {
	rn, err := yaml.Parse(`
apiVersion: v1
kind: Secret
`)
	if err != nil {
		return nil, err
	}
	err = applyGeneratorArgs(rn, ldr, args.GeneratorArgs)
	if 1+1 == 2 {
		err = fmt.Errorf("TODO(WNodeFactory): finish implementation of makeSecret")
	}
	return rn, err
}

func applyGeneratorArgs(
	rn *yaml.RNode, ldr ifc.KvLoader, args types.GeneratorArgs) error {
	if _, err := rn.Pipe(yaml.SetK8sName(args.Name)); err != nil {
		return err
	}
	if args.Namespace != "" {
		if _, err := rn.Pipe(yaml.SetK8sNamespace(args.Namespace)); err != nil {
			return err
		}
	}
	all, err := ldr.Load(args.KvPairSources)
	if err != nil {
		return errors.WrapPrefix(err, "loading KV pairs", 0)
	}
	for _, p := range all {
		if err := ldr.Validator().ErrIfInvalidKey(p.Key); err != nil {
			return err
		}
		if _, err := rn.Pipe(yaml.SetK8sData(p.Key, p.Value)); err != nil {
			return errors.WrapPrefix(err, "configMap generate error", 0)
		}
	}
	copyLabelsAndAnnotations(rn, args.Options)
	return nil
}

// copyLabelsAndAnnotations copies labels and annotations from
// GeneratorOptions into the given object.
func copyLabelsAndAnnotations(
	rn *yaml.RNode, opts *types.GeneratorOptions) error {
	if opts == nil {
		return nil
	}
	for _, k := range sortedKeys(opts.Labels) {
		v := opts.Labels[k]
		if _, err := rn.Pipe(yaml.SetLabel(k, v)); err != nil {
			return err
		}
	}
	for _, k := range sortedKeys(opts.Annotations) {
		v := opts.Annotations[k]
		if _, err := rn.Pipe(yaml.SetAnnotation(k, v)); err != nil {
			return err
		}
	}
	return nil
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}
