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
	"reflect"
	"strings"

	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kube-openapi/pkg/common"
	"sigs.k8s.io/kustomize/pkg/loader"
	"sigs.k8s.io/kustomize/pkg/transformers"
)

type pathConfigs struct {
	labelPathConfig          transformers.PathConfig
	annotationPathConfig     transformers.PathConfig
	prefixPathConfig         transformers.PathConfig
	namereferencePathConfigs []transformers.ReferencePathConfig
}

func (p *pathConfigs) addLabelPathConfig(config transformers.PathConfig) {
	if *(p.labelPathConfig.GroupVersionKind) == *config.GroupVersionKind {
		p.labelPathConfig.Path = append(p.labelPathConfig.Path, config.Path...)
	} else {
		p.labelPathConfig = config
	}
}

func (p *pathConfigs) addAnnotationPathConfig(config transformers.PathConfig) {
	if *(p.annotationPathConfig.GroupVersionKind) == *config.GroupVersionKind {
		p.annotationPathConfig.Path = append(p.labelPathConfig.Path, config.Path...)
	} else {
		p.annotationPathConfig = config
	}
}

func (p *pathConfigs) addNamereferencePathConfig(config transformers.ReferencePathConfig) {
	p.namereferencePathConfigs = transformers.MergeNameReferencePathConfigs(p.namereferencePathConfigs, config)
}

func (p *pathConfigs) addPrefixPathConfig(config transformers.PathConfig) {
	if *(p.prefixPathConfig.GroupVersionKind) == *config.GroupVersionKind {
		p.prefixPathConfig.Path = append(p.prefixPathConfig.Path, config.Path...)
	} else {
		p.prefixPathConfig = config
	}
}

// RegisterCRDs parse CRD schemas from paths and update various pathConfigs
func RegisterCRDs(loader loader.Loader, paths []string) error {
	var pathConfigs []pathConfigs
	for _, path := range paths {
		pathConfig, err := registerCRD(loader, path)
		if err != nil {
			return err
		}
		pathConfigs = append(pathConfigs, pathConfig...)
	}
	addPathConfigs(pathConfigs)
	return nil
}

// register CRD from one path
func registerCRD(loader loader.Loader, path string) ([]pathConfigs, error) {
	var result []pathConfigs
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
	for crd, gvk := range crds {
		crdPathConfigs := pathConfigs{}
		err = getCRDPathConfig(types, crd, crd, gvk, []string{}, &crdPathConfigs)
		if err != nil {
			return result, err
		}
		if !reflect.DeepEqual(crdPathConfigs, pathConfigs{}) {
			result = append(result, crdPathConfigs)
		}
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
func getCRDPathConfig(types map[string]common.OpenAPIDefinition, atype string, crd string, gvk schema.GroupVersionKind,
	path []string, configs *pathConfigs) error {
	if _, ok := types[crd]; !ok {
		return nil
	}

	for propname, property := range types[atype].Schema.SchemaProps.Properties {
		_, annotate := property.Extensions.GetString(Annotation)
		if annotate {
			configs.addAnnotationPathConfig(
				transformers.PathConfig{
					CreateIfNotPresent: false,
					GroupVersionKind:   &gvk,
					Path:               append(path, propname),
				},
			)
		}
		_, label := property.Extensions.GetString(LabelSelector)
		if label {
			configs.addLabelPathConfig(
				transformers.PathConfig{
					CreateIfNotPresent: false,
					GroupVersionKind:   &gvk,
					Path:               append(path, propname),
				},
			)
		}
		_, identity := property.Extensions.GetString(Identity)
		if identity {
			configs.addPrefixPathConfig(
				transformers.PathConfig{
					CreateIfNotPresent: false,
					GroupVersionKind:   &gvk,
					Path:               append(path, propname),
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
				configs.addNamereferencePathConfig(transformers.NewReferencePathConfig(
					schema.GroupVersionKind{Kind: kind, Version: version},
					[]transformers.PathConfig{
						{CreateIfNotPresent: false,
							GroupVersionKind: &gvk,
							Path:             append(path, propname, nameKey),
						}}))

			}
		}

		if property.Ref.GetURL() != nil {
			getCRDPathConfig(types, property.Ref.String(), crd, gvk, append(path, propname), configs)
		}
	}
	return nil
}

// addPathConfigs add extra path configs to the default ones
func addPathConfigs(p []pathConfigs) {
	for _, pc := range p {
		transformers.AddLabelsPathConfigs(pc.labelPathConfig)
		transformers.AddAnnotationsPathConfigs(pc.annotationPathConfig)
		transformers.AddNameReferencePathConfigs(pc.namereferencePathConfigs)
		transformers.AddPrefixPathConfigs(pc.prefixPathConfig)
	}
}
