module sigs.k8s.io/kustomize/plugin/builtin/helmchartinflationgenerator

go 1.16

require (
	github.com/imdario/mergo v0.3.5
	github.com/pkg/errors v0.9.1
	sigs.k8s.io/kustomize/api v0.8.9
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/kustomize/api => ../../../api
