module sigs.k8s.io/kustomize/plugin/someteam.example.com/v1/dateprefixer

go 1.16

require (
	github.com/pkg/errors v0.9.1
	sigs.k8s.io/kustomize/api v0.8.5
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/kustomize/kyaml => ../../../../kyaml
