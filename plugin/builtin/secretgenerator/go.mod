module sigs.k8s.io/kustomize/plugin/builtin/secretgenerator

go 1.14

require (
	sigs.k8s.io/kustomize/api v0.6.2
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/kustomize/api v0.6.2 => ../../../api
