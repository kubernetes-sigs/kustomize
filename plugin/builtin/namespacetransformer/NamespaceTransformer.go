// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

//go:generate go run sigs.k8s.io/kustomize/v3/cmd/pluginator
package main

import (
	"sigs.k8s.io/kustomize/v3/pkg/gvk"
	"sigs.k8s.io/kustomize/v3/pkg/ifc"
	"sigs.k8s.io/kustomize/v3/pkg/resid"
	"sigs.k8s.io/kustomize/v3/pkg/resmap"
	"sigs.k8s.io/kustomize/v3/pkg/resource"
	"sigs.k8s.io/kustomize/v3/pkg/transformers"
	"sigs.k8s.io/kustomize/v3/pkg/transformers/config"
	"sigs.k8s.io/kustomize/v3/pkg/types"
	"sigs.k8s.io/yaml"
)

// Change or set the namespace of non-cluster level resources.
type plugin struct {
	types.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	FieldSpecs       []config.FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
}

//noinspection GoUnusedGlobalVariable
var KustomizePlugin plugin

func (p *plugin) Config(
	ldr ifc.Loader, rf *resmap.Factory, c []byte) (err error) {
	p.Namespace = ""
	p.FieldSpecs = nil
	return yaml.Unmarshal(c, p)
}

func (p *plugin) Transform(m resmap.ResMap) error {
	if len(p.Namespace) == 0 {
		return nil
	}
	for _, r := range m.Resources() {
		id := r.OrgId()
		fs, ok := p.isSelected(id)
		if !ok {
			continue
		}
		if len(r.Map()) == 0 {
			// Don't mutate empty objects?
			continue
		}
		if doIt(id, fs) {
			if err := p.changeNamespace(r, fs); err != nil {
				return err
			}
		}
	}
	p.updateClusterRoleBinding(m)
	p.updateServiceReference(m)
	return nil
}

const metaNamespace = "metadata/namespace"

// Special casing metadata.namespace since
// all objects have it, even "ClusterKind" objects
// that don't exist in a namespace (the Namespace
// object itself doesn't live in a namespace).
func doIt(id resid.ResId, fs *config.FieldSpec) bool {
	return fs.Path != metaNamespace ||
		(fs.Path == metaNamespace && id.IsNamespaceableKind())
}

func (p *plugin) changeNamespace(
	r *resource.Resource, fs *config.FieldSpec) error {
	return transformers.MutateField(
		r.Map(), fs.PathSlice(), fs.CreateIfNotPresent,
		func(_ interface{}) (interface{}, error) {
			return p.Namespace, nil
		})
}

func (p *plugin) isSelected(
	id resid.ResId) (*config.FieldSpec, bool) {
	for _, fs := range p.FieldSpecs {
		if id.IsSelected(&fs.Gvk) {
			return &fs, true
		}
	}
	return nil, false
}

func (p *plugin) updateClusterRoleBinding(m resmap.ResMap) {
	srvAccount := gvk.Gvk{Version: "v1", Kind: "ServiceAccount"}
	saMap := map[string]bool{}
	for _, id := range m.AllIds() {
		if id.Gvk.Equals(srvAccount) {
			saMap[id.Name] = true
		}
	}

	for _, res := range m.Resources() {
		if res.OrgId().Kind != "ClusterRoleBinding" &&
			res.OrgId().Kind != "RoleBinding" {
			continue
		}
		objMap := res.Map()
		subjects, ok := objMap["subjects"].([]interface{})
		if subjects == nil || !ok {
			continue
		}
		for i := range subjects {
			subject := subjects[i].(map[string]interface{})
			kind, foundK := subject["kind"]
			name, foundN := subject["name"]
			if !foundK || !foundN || kind.(string) != srvAccount.Kind {
				continue
			}
			// a ServiceAccount named “default” exists in every active namespace
			if name.(string) == "default" || saMap[name.(string)] {
				subject := subjects[i].(map[string]interface{})
				transformers.MutateField(
					subject, []string{"namespace"},
					true, func(_ interface{}) (interface{}, error) {
						return p.Namespace, nil
					})
				subjects[i] = subject
			}
		}
		objMap["subjects"] = subjects
	}
}

func (p *plugin) updateServiceReference(m resmap.ResMap) {
	svc := gvk.Gvk{Version: "v1", Kind: "Service"}
	svcMap := map[string]bool{}
	for _, id := range m.AllIds() {
		if id.Gvk.Equals(svc) {
			svcMap[id.Name] = true
		}
	}

	for _, res := range m.Resources() {
		if res.OrgId().Kind != "ValidatingWebhookConfiguration" &&
			res.OrgId().Kind != "MutatingWebhookConfiguration" {
			continue
		}
		objMap := res.Map()
		webhooks, ok := objMap["webhooks"].([]interface{})
		if webhooks == nil || !ok {
			continue
		}
		for i := range webhooks {
			webhook := webhooks[i].(map[string]interface{})
			transformers.MutateField(
				webhook, []string{"clientConfig", "service"},
				false, func(obj interface{}) (interface{}, error) {
					svc := obj.(map[string]interface{})
					svcName, foundN := svc["name"]
					if foundN && svcMap[svcName.(string)] {
						svc["namespace"] = p.Namespace
					}
					return svc, nil
				})
			webhooks[i] = webhook
		}
		objMap["webhooks"] = webhooks
	}
}
