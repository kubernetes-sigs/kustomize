module sigs.k8s.io/kustomize/plugin/builtin/replicacounttransformer

go 1.14

require (
	sigs.k8s.io/kustomize/api v0.5.1
	sigs.k8s.io/kustomize/kyaml v0.7.0
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/kustomize/api v0.5.1 => ../../../api

replace sigs.k8s.io/kustomize/kyaml v0.7.0 => ../../../kyaml
