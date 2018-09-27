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

// Package crds read in files for CRD schemas and parse annotations from it
package crds

import (
	"encoding/json"
	"strings"

	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kube-openapi/pkg/common"
	"sigs.k8s.io/kustomize/pkg/gvk"
	"sigs.k8s.io/kustomize/pkg/loader"
	"sigs.k8s.io/kustomize/pkg/transformerconfig"
)

// RegisterCRDs parse CRD schemas from paths and update various pathConfigs
func RegisterCRDs(loader loader.Loader, paths []string) (*transformerconfig.TransformerConfig, error) {
	pathConfigs := transformerconfig.MakeEmptyTransformerConfig()
	for _, path := range paths {
		pathConfig, err := registerCRD(loader, path)
		if err != nil {
			return nil, err
		}
		pathConfigs = pathConfigs.Merge(pathConfig)
	}
	return pathConfigs, nil
}

// register CRD from one path
func registerCRD(loader loader.Loader, path string) (*transformerconfig.TransformerConfig, error) {
	result := transformerconfig.MakeEmptyTransformerConfig()
	content, err := loader.Load(path)
	if err != nil {
		return result, err
	}

	var types map[string]common.OpenAPIDefinition
	if content[0] == '{' {
		err = json.Unmarshal(content, &types)
		if err != nil {
			return nil, err
		}
	} else {
		err = yaml.Unmarshal(content, &types)
		if err != nil {
			return nil, err
		}
	}

	crds := getCRDs(types)
	for crd, k := range crds {
		crdPathConfigs := transformerconfig.MakeEmptyTransformerConfig()
		err = getCRDPathConfig(
			types, crd, crd, gvk.FromSchemaGvk(k), []string{}, crdPathConfigs)
		if err != nil {
			return result, err
		}
		result = result.Merge(crdPathConfigs)
	}

	return result, nil
}

// getCRDs get all CRD types
func getCRDs(types map[string]common.OpenAPIDefinition) map[string]schema.GroupVersionKind {
	crds := map[string]schema.GroupVersionKind{}

	for typename, t := range types {
		properties := t.Schema.SchemaProps.Properties
		_, foundKind := properties["kind"]
		_, foundAPIVersion := properties["apiVersion"]
		_, foundMetadata := properties["metadata"]
		if foundKind && foundAPIVersion && foundMetadata {
			// TODO: Get Group and Version for CRD from the openAPI definition once
			// "x-kubernetes-group-version-kind" is available in CRD
			kind := strings.Split(typename, ".")[len(strings.Split(typename, "."))-1]
			crds[typename] = schema.GroupVersionKind{Kind: kind}
		}
	}
	return crds
}

// getCRDPathConfig gets pathConfigs for one CRD recursively
func getCRDPathConfig(
	types map[string]common.OpenAPIDefinition, atype string, crd string, in gvk.Gvk,
	path []string, configs *transformerconfig.TransformerConfig) error {
	if _, ok := types[crd]; !ok {
		return nil
	}

	for propname, property := range types[atype].Schema.SchemaProps.Properties {
		_, annotate := property.Extensions.GetString(Annotation)
		if annotate {
			configs.AddAnnotationPathConfig(
				transformerconfig.PathConfig{
					CreateIfNotPresent: false,
					Gvk:                in,
					Path:               strings.Join(append(path, propname), "/"),
				},
			)
		}
		_, label := property.Extensions.GetString(LabelSelector)
		if label {
			configs.AddLabelPathConfig(
				transformerconfig.PathConfig{
					CreateIfNotPresent: false,
					Gvk:                in,
					Path:               strings.Join(append(path, propname), "/"),
				},
			)
		}
		_, identity := property.Extensions.GetString(Identity)
		if identity {
			configs.AddPrefixPathConfig(
				transformerconfig.PathConfig{
					CreateIfNotPresent: false,
					Gvk:                in,
					Path:               strings.Join(append(path, propname), "/"),
				},
			)
		}
		version, ok := property.Extensions.GetString(Version)
		if ok {
			kind, ok := property.Extensions.GetString(Kind)
			if ok {
				nameKey, ok := property.Extensions.GetString(NameKey)
				if !ok {
					nameKey = "name"
				}
				configs.AddNamereferencePathConfig(transformerconfig.ReferencePathConfig{
					Gvk: gvk.Gvk{Kind: kind, Version: version},
					PathConfigs: []transformerconfig.PathConfig{
						{
							CreateIfNotPresent: false,
							Gvk:                in,
							Path:               strings.Join(append(path, propname, nameKey), "/"),
						},
					},
				})
			}
		}

		if property.Ref.GetURL() != nil {
			getCRDPathConfig(types, property.Ref.String(), crd, in, append(path, propname), configs)
		}
	}
	return nil
}
