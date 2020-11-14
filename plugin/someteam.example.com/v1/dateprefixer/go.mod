module sigs.k8s.io/kustomize/plugin/someteam.example.com/v1/dateprefixer

go 1.14

require (
	github.com/pkg/errors v0.8.1
	sigs.k8s.io/kustomize/api v0.6.5
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/kustomize/api v0.6.5 => ../../../../api
replace sigs.k8s.io/kustomize/kyaml v0.9.4 => ../../../../kyaml
