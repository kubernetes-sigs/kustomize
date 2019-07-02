// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package resource

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	jsonpatch "github.com/evanphx/json-patch"
	"sigs.k8s.io/yaml"

	"sigs.k8s.io/kustomize/v3/internal/kusterr"
	"sigs.k8s.io/kustomize/v3/pkg/ifc"
	"sigs.k8s.io/kustomize/v3/pkg/types"
)

// Factory makes instances of Resource.
type Factory struct {
	kf ifc.KunstructuredFactory
}

// NewFactory makes an instance of Factory.
func NewFactory(kf ifc.KunstructuredFactory) *Factory {
	return &Factory{kf: kf}
}

func (rf *Factory) Hasher() ifc.KunstructuredHasher {
	return rf.kf.Hasher()
}

// FromMap returns a new instance of Resource.
func (rf *Factory) FromMap(m map[string]interface{}) *Resource {
	return rf.makeOne(rf.kf.FromMap(m), nil)
}

// FromMapWithName returns a new instance with the given "original" name.
func (rf *Factory) FromMapWithName(n string, m map[string]interface{}) *Resource {
	return rf.makeOne(rf.kf.FromMap(m), nil).setOriginalName(n)
}

// FromMapWithNamespace returns a new instance with the given "original" namespace.
func (rf *Factory) FromMapWithNamespace(n string, m map[string]interface{}) *Resource {
	return rf.makeOne(rf.kf.FromMap(m), nil).setOriginalNs(n)
}

// FromMapAndOption returns a new instance of Resource with given options.
func (rf *Factory) FromMapAndOption(
	m map[string]interface{}, args *types.GeneratorArgs, option *types.GeneratorOptions) *Resource {
	return rf.makeOne(rf.kf.FromMap(m), types.NewGenArgs(args, option))
}

// FromKunstructured returns a new instance of Resource.
func (rf *Factory) FromKunstructured(u ifc.Kunstructured) *Resource {
	return rf.makeOne(u, nil)
}

// makeOne returns a new instance of Resource.
func (rf *Factory) makeOne(
	u ifc.Kunstructured, o *types.GenArgs) *Resource {
	if u == nil {
		log.Fatal("unstruct ifc must not be null")
	}
	if o == nil {
		o = types.NewGenArgs(nil, nil)
	}
	r := &Resource{
		Kunstructured: u,
		options:       o,
	}
	return r.setOriginalName(r.GetName()).setOriginalNs(r.GetNamespace())
}

// SliceFromPatches returns a slice of resources given a patch path
// slice from a kustomization file.
func (rf *Factory) SliceFromPatches(
	ldr ifc.Loader, paths []types.PatchStrategicMerge) ([]*Resource, error) {
	var result []*Resource
	for _, path := range paths {
		content, err := ldr.Load(string(path))
		if err != nil {
			return nil, err
		}
		res, err := rf.SliceFromBytes(content)
		if err != nil {
			return nil, kusterr.Handler(err, string(path))
		}
		result = append(result, res...)
	}
	return result, nil
}

// PatchFromPatchArg returns a Strategic Merge Patch or a Json patch
// from a Patch argument
func (rf *Factory) PatchFromPatchArg(
	ldr ifc.Loader, patchArg types.Patch) (*Resource, jsonpatch.Patch, error) {
	if patchArg.Patch != "" && patchArg.Path != "" {
		return nil, nil, fmt.Errorf(
			"patch and path can't be set at the same time %v", patchArg)
	}
	var err error
	var in []byte
	if patchArg.Path != "" {
		in, err = ldr.Load(patchArg.Path)
		if err != nil {
			return nil, nil, err
		}
	}
	if patchArg.Patch != "" {
		in = []byte(patchArg.Patch)
	}

	patchSM, errSM := rf.FromBytes(in)
	patchJson, errJson := rf.JsonPatchFromBytes(in)
	if errSM != nil && errJson != nil {
		return nil, nil, fmt.Errorf(
			"unable to get either a Strategic Merge Patch or JSON patch 6902 from %v", patchArg)
	}
	if errSM == nil && errJson != nil {
		return patchSM, nil, nil
	}
	if errJson == nil && errSM != nil {
		return nil, patchJson, nil
	}
	if patchSM != nil && patchJson != nil {
		return nil, nil, fmt.Errorf(
			"a patch can't be both a Strategic Merge Patch and JSON patch 6902 %v", patchArg)
	}
	return nil, nil, nil
}

// JsonPatchFromBytes loads a Json 6902 patch from
// a bytes input
func (rf Factory) JsonPatchFromBytes(
	in []byte) (jsonpatch.Patch, error) {
	ops := string(in)
	if ops == "" {
		return nil, fmt.Errorf("empty json patch operations")
	}

	if ops[0] != '[' {
		jsonOps, err := yaml.YAMLToJSON(in)
		if err != nil {
			return nil, err
		}
		ops = string(jsonOps)
	}
	return jsonpatch.DecodePatch([]byte(ops))
}

// FromBytes unmarshals bytes into one Resource.
func (rf *Factory) FromBytes(in []byte) (*Resource, error) {
	result, err := rf.SliceFromBytes(in)
	if err != nil {
		return nil, err
	}
	if len(result) != 1 {
		return nil, fmt.Errorf(
			"expected 1 resource, found %d in %v", len(result), in)
	}
	return result[0], nil
}

// SliceFromBytes unmarshals bytes into a Resource slice.
func (rf *Factory) SliceFromBytes(in []byte) ([]*Resource, error) {
	kunStructs, err := rf.kf.SliceFromBytes(in)
	if err != nil {
		return nil, err
	}
	var result []*Resource
	for len(kunStructs) > 0 {
		u := kunStructs[0]
		kunStructs = kunStructs[1:]
		if strings.HasSuffix(u.GetKind(), "List") {
			items := u.Map()["items"]
			itemsSlice, ok := items.([]interface{})
			if !ok {
				if items == nil {
					// an empty list
					continue
				}
				return nil, fmt.Errorf("items in List is type %T, expected array", items)
			}
			for _, item := range itemsSlice {
				itemJSON, err := json.Marshal(item)
				if err != nil {
					return nil, err
				}
				innerU, err := rf.kf.SliceFromBytes(itemJSON)
				if err != nil {
					return nil, err
				}
				// append innerU to kunStructs so nested Lists can be handled
				kunStructs = append(kunStructs, innerU...)
			}
		} else {
			result = append(result, rf.FromKunstructured(u))
		}
	}
	return result, nil
}

// MakeConfigMap makes an instance of Resource for ConfigMap
func (rf *Factory) MakeConfigMap(
	ldr ifc.Loader,
	options *types.GeneratorOptions,
	args *types.ConfigMapArgs) (*Resource, error) {
	u, err := rf.kf.MakeConfigMap(ldr, options, args)
	if err != nil {
		return nil, err
	}
	return rf.makeOne(
		u,
		types.NewGenArgs(
			&types.GeneratorArgs{Behavior: args.Behavior},
			options)), nil
}

// MakeSecret makes an instance of Resource for Secret
func (rf *Factory) MakeSecret(
	ldr ifc.Loader,
	options *types.GeneratorOptions,
	args *types.SecretArgs) (*Resource, error) {
	u, err := rf.kf.MakeSecret(ldr, options, args)
	if err != nil {
		return nil, err
	}
	return rf.makeOne(
		u,
		types.NewGenArgs(
			&types.GeneratorArgs{Behavior: args.Behavior},
			options)), nil
}
