module sigs.k8s.io/kustomize/plugin/builtin/labeltransformer

go 1.13

require (
	sigs.k8s.io/kustomize/api v0.3.1
	sigs.k8s.io/kustomize/kyaml v0.1.5
	sigs.k8s.io/yaml v1.1.0
)

replace sigs.k8s.io/kustomize/api v0.3.1 => ../../../api
