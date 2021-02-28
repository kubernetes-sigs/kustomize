module sigs.k8s.io/kustomize/plugin/builtin/legacyordertransformer

go 1.15

require (
	github.com/pkg/errors v0.9.1
	sigs.k8s.io/kustomize/api v0.8.4
)

replace sigs.k8s.io/kustomize/kyaml => ../../../kyaml

replace sigs.k8s.io/kustomize/api => ../../../api
