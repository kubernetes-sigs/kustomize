module sigs.k8s.io/kustomize/plugin/builtin/replacementtransformer

go 1.16

require (
	sigs.k8s.io/kustomize/api v0.8.9
	sigs.k8s.io/yaml v1.3.0
)

replace sigs.k8s.io/kustomize/kyaml => ../../../kyaml

replace sigs.k8s.io/kustomize/api => ../../../api
