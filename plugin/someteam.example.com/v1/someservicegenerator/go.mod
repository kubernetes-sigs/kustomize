module sigs.k8s.io/kustomize/plugin/someteam.example.com/v1/someservicegenerator

go 1.13

require (
	sigs.k8s.io/kustomize/api v0.2.0
	sigs.k8s.io/yaml v1.1.0
)

replace sigs.k8s.io/kustomize/api v0.2.0 => ../../../../api
