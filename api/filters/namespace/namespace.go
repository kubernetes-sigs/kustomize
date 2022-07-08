// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package namespace

import (
	"sigs.k8s.io/kustomize/api/filters/filtersutil"
	"sigs.k8s.io/kustomize/api/filters/fsslice"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/resid"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type Filter struct {
	// Namespace is the namespace to apply to the inputs
	Namespace string `yaml:"namespace,omitempty"`

	// FsSlice contains the FieldSpecs to locate the namespace field
	FsSlice types.FsSlice `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`

	// SkipExisting means only blank namespace fields will be set
	SkipExisting bool `json:"skipExisting" yaml:"skipExisting"`

	trackableSetter filtersutil.TrackableSetter
}

var _ kio.Filter = Filter{}
var _ kio.TrackableFilter = &Filter{}

// WithMutationTracker registers a callback which will be invoked each time a field is mutated
func (ns *Filter) WithMutationTracker(callback func(key, value, tag string, node *yaml.RNode)) {
	ns.trackableSetter.WithMutationTracker(callback)
}

func (ns Filter) Filter(nodes []*yaml.RNode) ([]*yaml.RNode, error) {
	return kio.FilterAll(yaml.FilterFunc(ns.run)).Filter(nodes)
}

// Run runs the filter on a single node rather than a slice
func (ns Filter) run(node *yaml.RNode) (*yaml.RNode, error) {
	// Special handling for metadata.namespace -- :(
	// never let SetEntry handle metadata.namespace--it will incorrectly include cluster-scoped resources
	ns.FsSlice = ns.removeMetaNamespaceFieldSpecs(ns.FsSlice)
	gvk := resid.GvkFromNode(node)
	if err := ns.metaNamespaceHack(node, gvk); err != nil {
		return nil, err
	}

	// Special handling for (cluster) role binding -- :(
	if isRoleBinding(gvk.Kind) {
		ns.FsSlice = ns.removeRoleBindingFieldSpecs(ns.FsSlice)
		if err := ns.roleBindingHack(node, gvk); err != nil {
			return nil, err
		}
	}

	// transformations based on data -- :)
	err := node.PipeE(fsslice.Filter{
		FsSlice:    ns.FsSlice,
		SetValue:   ns.fieldSetter(),
		CreateKind: yaml.ScalarNode, // Namespace is a ScalarNode
		CreateTag:  yaml.NodeTagString,
	})
	return node, err
}

// metaNamespaceHack is a hack for implementing the namespace transform
// for the metadata.namespace field on namespace scoped resources.
func (ns Filter) metaNamespaceHack(obj *yaml.RNode, gvk resid.Gvk) error {
	if gvk.IsClusterScoped() {
		return nil
	}
	f := fsslice.Filter{
		FsSlice: []types.FieldSpec{
			{Path: types.MetadataNamespacePath, CreateIfNotPresent: true},
		},
		SetValue:   ns.fieldSetter(),
		CreateKind: yaml.ScalarNode, // Namespace is a ScalarNode
	}
	_, err := f.Filter(obj)
	return err
}

// roleBindingHack is a hack for implementing the transformer's DefaultSubjectsOnly mode
// for RoleBinding and ClusterRoleBinding resource types.
// In this mode, RoleBinding and ClusterRoleBinding have namespace set on
// elements of the "subjects" field if and only if the subject elements
// "name" is "default".  Otherwise the namespace is not set.
//
// Example:
//
// kind: RoleBinding
// subjects:
// - name: "default" # this will have the namespace set
//   ...
// - name: "something-else" # this will not have the namespace set
//   ...
func (ns Filter) roleBindingHack(obj *yaml.RNode, gvk resid.Gvk) error {
	if !isRoleBinding(gvk.Kind) {
		return nil
	}

	// Lookup the namespace field on all elements.
	obj, err := obj.Pipe(yaml.Lookup(subjectsField))
	if err != nil || yaml.IsMissingOrNull(obj) {
		return err
	}

	// add the namespace to each "subject" with name: default
	err = obj.VisitElements(func(o *yaml.RNode) error {
		// The only case we need to force the namespace
		// if for the "service account". "default" is
		// kind of hardcoded here for right now.
		name, err := o.Pipe(
			yaml.Lookup("name"), yaml.Match("default"),
		)
		if err != nil || yaml.IsMissingOrNull(name) {
			return err
		}

		// set the namespace for the default account
		node, err := o.Pipe(
			yaml.LookupCreate(yaml.ScalarNode, "namespace"),
		)
		if err != nil {
			return err
		}

		return ns.fieldSetter()(node)
	})

	return err
}

func isRoleBinding(kind string) bool {
	switch kind {
	case roleBindingKind, clusterRoleBindingKind:
		return true
	default:
		return false
	}
}

// removeRoleBindingFieldSpecs removes from the list fieldspecs that
// have hardcoded implementations
func (ns Filter) removeRoleBindingFieldSpecs(fs types.FsSlice) types.FsSlice {
	var val types.FsSlice
	for i := range fs {
		if isRoleBinding(fs[i].Kind) && fs[i].Path == subjectsField {
			continue
		}
		val = append(val, fs[i])
	}
	return val
}

func (ns Filter) removeMetaNamespaceFieldSpecs(fs types.FsSlice) types.FsSlice {
	var val types.FsSlice
	for i := range fs {
		if fs[i].Path == types.MetadataNamespacePath {
			continue
		}
		val = append(val, fs[i])
	}
	return val
}

func (ns *Filter) fieldSetter() filtersutil.SetFn {
	if ns.SkipExisting {
		return ns.trackableSetter.SetEntryIfEmpty("", ns.Namespace, yaml.NodeTagString)
	}
	return ns.trackableSetter.SetEntry("", ns.Namespace, yaml.NodeTagString)
}

const (
	subjectsField          = "subjects"
	roleBindingKind        = "RoleBinding"
	clusterRoleBindingKind = "ClusterRoleBinding"
)
