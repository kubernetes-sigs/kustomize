module sigs.k8s.io/kustomize/plugin/someteam.example.com/configmappatchjsontransformer

go 1.14

require (
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/pkg/errors v0.8.1
	sigs.k8s.io/kustomize/api v0.5.1
	sigs.k8s.io/kustomize/kyaml v0.7.1
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/kustomize/api v0.5.1 => ../../../../api

replace sigs.k8s.io/kustomize/kyaml v0.7.1 => ../../../../kyaml
