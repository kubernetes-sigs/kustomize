module sigs.k8s.io/kustomize/plugin/builtin/annotationstransformer

go 1.15

require (
	sigs.k8s.io/kustomize/api v0.7.0
	sigs.k8s.io/kustomize/kyaml v0.10.3
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/kustomize/api v0.7.0 => ../../../api

replace sigs.k8s.io/kustomize/kyaml v0.10.3 => ../../../kyaml
