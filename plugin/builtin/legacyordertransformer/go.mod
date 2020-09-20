module sigs.k8s.io/kustomize/plugin/builtin/legacyordertransformer

go 1.14

require (
	github.com/pkg/errors v0.8.1
	sigs.k8s.io/kustomize/api v0.6.2
)

replace sigs.k8s.io/kustomize/api v0.6.2 => ../../../api
