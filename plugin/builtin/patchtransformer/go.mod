module sigs.k8s.io/kustomize/plugin/builtin/patchtransformer

go 1.16

require (
	github.com/evanphx/json-patch v4.11.0+incompatible
	sigs.k8s.io/kustomize/api v0.8.9
	sigs.k8s.io/kustomize/kyaml v0.13.0
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/kustomize/api => ../../../api

replace sigs.k8s.io/kustomize/kyaml => ../../../kyaml
