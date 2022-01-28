module sigs.k8s.io/kustomize/plugin/builtin/legacyordertransformer

go 1.16

require (
	github.com/pkg/errors v0.9.1
	sigs.k8s.io/kustomize/api v0.8.9
	sigs.k8s.io/kustomize/kyaml v0.13.0 // indirect
)

replace sigs.k8s.io/kustomize/api => ../../../api

replace sigs.k8s.io/kustomize/kyaml => ../../../kyaml
