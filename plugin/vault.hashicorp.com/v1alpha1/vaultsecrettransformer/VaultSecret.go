package main

import (
	"sigs.k8s.io/kustomize/api/resid"
	"sigs.k8s.io/kustomize/api/types"
)

var VaultSecretGvk = resid.Gvk{
	Group:   "vault.hashicorp.com",
	Version: "v1alpha1",
	Kind:    "VaultSecret",
}

type VaultSecret struct {
	types.TypeMeta   `json:",inline"`
	types.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec             VaultSecretSpec `json:"spec,omitempty" yaml:"spec,omitempty"`
}

type VaultSecretSpec struct {
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
}
