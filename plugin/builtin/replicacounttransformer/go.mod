module sigs.k8s.io/kustomize/plugin/builtin/replicacounttransformer

go 1.15

require (
	sigs.k8s.io/kustomize/api v0.7.2
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/kustomize/kyaml => ../../../kyaml

replace sigs.k8s.io/kustomize/api => ../../../api
