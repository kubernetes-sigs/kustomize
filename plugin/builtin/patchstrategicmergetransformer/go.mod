module sigs.k8s.io/kustomize/plugin/builtin/patchstrategicmergetransformer

go 1.16

require (
	github.com/stretchr/testify v1.7.0
	sigs.k8s.io/kustomize/api v0.8.9
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/kustomize/api => ../../../api

replace sigs.k8s.io/kustomize/kyaml => ../../../kyaml
