module sigs.k8s.io/kustomize/plugin/builtin/labeltransformer

go 1.13

require (
	sigs.k8s.io/kustomize/api v0.0.0
	sigs.k8s.io/kustomize/kyaml v0.1.10
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/kustomize/api v0.0.0 => ../../../api
