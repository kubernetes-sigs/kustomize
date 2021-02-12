module sigs.k8s.io/kustomize/plugin/builtin/prefixsuffixtransformer

go 1.15

require (
	sigs.k8s.io/kustomize/api v0.8.1
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/kustomize/kyaml => ../../../kyaml

replace sigs.k8s.io/kustomize/api => ../../../api
