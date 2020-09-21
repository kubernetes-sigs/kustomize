module sigs.k8s.io/kustomize/plugin/builtin/patchtransformer

go 1.14

require (
	github.com/evanphx/json-patch v4.5.0+incompatible
	sigs.k8s.io/kustomize/api v0.6.2
	sigs.k8s.io/kustomize/kyaml v0.8.1
	github.com/pkg/errors v0.8.1
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/kustomize/kyaml v0.8.1 => ../../../kyaml

replace sigs.k8s.io/kustomize/api v0.6.2 => ../../../api
