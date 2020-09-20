module sigs.k8s.io/kustomize/plugin/builtin/valueaddtransformer

go 1.14

require (
	sigs.k8s.io/kustomize/api v0.6.2
	sigs.k8s.io/kustomize/kyaml v0.8.1
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/kustomize/kyaml v0.8.1 => ../../../kyaml

replace sigs.k8s.io/kustomize/api v0.6.2 => ../../../api
