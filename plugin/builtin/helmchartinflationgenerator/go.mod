module sigs.k8s.io/kustomize/plugin/builtin/helmchartinflationgenerator

go 1.15

require (
	github.com/imdario/mergo v0.3.5
	github.com/pkg/errors v0.8.1
	sigs.k8s.io/kustomize/api v0.7.1
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/kustomize/api => ../../../api
