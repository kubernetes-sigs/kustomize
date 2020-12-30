module sigs.k8s.io/kustomize/plugin/builtin/patchstrategicmergetransformer

go 1.15

require (
	github.com/stretchr/testify v1.4.0
	sigs.k8s.io/kustomize/api v0.7.1
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/kustomize/kyaml => ../../../kyaml

replace sigs.k8s.io/kustomize/api => ../../../api
