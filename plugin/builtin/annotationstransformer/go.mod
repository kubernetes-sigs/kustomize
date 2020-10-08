module sigs.k8s.io/kustomize/plugin/builtin/annotationstransformer

go 1.14

require (
	sigs.k8s.io/kustomize/api v0.6.3
	sigs.k8s.io/kustomize/kyaml v0.9.1
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/kustomize/api v0.6.3 => ../../../api

replace sigs.k8s.io/kustomize/kyaml v0.9.1 => ../../../kyaml
