module sigs.k8s.io/kustomize/plugin/builtin/labeltransformer

go 1.14

require (
	sigs.k8s.io/kustomize/api v0.4.0
	sigs.k8s.io/kustomize/kyaml v0.3.2
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/kustomize/api v0.4.0 => ../../../api
