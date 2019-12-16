// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package fieldspec

import (
	"sigs.k8s.io/kustomize/functions/config.kubernetes.io/kustomize/util"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// KustomizeAnnotationsFilter sets Annotations on all Resources in a package, including sub-field
// annotations -- e.g. spec.template.metadata.annotations.
// Overrides existing annotations iff the keys match.
type KustomizeAnnotationsFilter struct {
	// commonAnnotations are the annotations to set
	Annotations map[string]*string `yaml:"commonAnnotations,omitempty"`

	kustomizeName string
}

func (ns *KustomizeAnnotationsFilter) SetKustomizeName(n string) *KustomizeAnnotationsFilter {
	ns.kustomizeName = n
	return ns
}

var _ kio.Filter = KustomizeAnnotationsFilter{}

func (af KustomizeAnnotationsFilter) Filter(input []*yaml.RNode) ([]*yaml.RNode, error) {
	for k, v := range af.Annotations {
		f := af.AnnotationFilter(k, v)
		_, err := kio.FilterAll(f).Filter(input)
		if err != nil {
			return nil, err
		}
	}
	return input, nil
}

var annotationFieldSpecList *FieldSpecList

func getAnnotationReferenceFieldSpecs() *FieldSpecList {
	if annotationFieldSpecList != nil {
		return annotationFieldSpecList
	}
	annotationFieldSpecList = &FieldSpecList{}
	err := yaml.Unmarshal([]byte(commonAnnotationFieldSpecs), &annotationFieldSpecList)
	if err != nil {
		panic(err)
	}
	return annotationFieldSpecList
}

func (af KustomizeAnnotationsFilter) AnnotationFilter(key string, value *string) *FieldSpecListFilter {
	if value != nil {
		return &FieldSpecListFilter{
			FieldSpecList: *getAnnotationReferenceFieldSpecs(),
			SetValue: func(node *yaml.RNode) error {
				v := yaml.NewScalarRNode(*value)
				if err := util.SetSetter(v, af.kustomizeName); err != nil {
					return err
				}
				return node.PipeE(yaml.FieldSetter{Name: key, Value: v})
			},
			CreateKind: yaml.MappingNode,
		}
	}

	return &FieldSpecListFilter{
		FieldSpecList: *getAnnotationReferenceFieldSpecs(),
		SetValue: func(node *yaml.RNode) error {
			_, err := node.Pipe(yaml.FieldClearer{Name: key})
			return err
		},
		CreateKind: yaml.MappingNode,
	}
}

const commonAnnotationFieldSpecs = `
items:

# duck-type supported annotations
#
- path: metadata/annotations
  create: true
- path: spec/template/metadata/annotations
  create: false

# non-duck-type supported annotations
#
- path: spec/template/metadata/annotations
  create: true
  version: v1
  kind: ReplicationController

- path: spec/template/metadata/annotations
  create: true
  kind: Deployment

- path: spec/template/metadata/annotations
  create: true
  kind: ReplicaSet

- path: spec/template/metadata/annotations
  create: true
  kind: DaemonSet

- path: spec/template/metadata/annotations
  create: true
  kind: StatefulSet

- path: spec/template/metadata/annotations
  create: true
  group: batch
  kind: Job

- path: spec/jobTemplate/metadata/annotations
  create: true
  group: batch
  kind: CronJob

- path: spec/jobTemplate/spec/template/metadata/annotations
  create: true
  group: batch
  kind: CronJob

`
