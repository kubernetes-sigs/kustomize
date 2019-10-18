module sigs.k8s.io/kustomize/plugin/someteam.example.com/v1/replacementtransformer

go 1.13

require (
	github.com/emicklei/go-restful v2.9.6+incompatible // indirect
	github.com/googleapis/gnostic v0.3.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	sigs.k8s.io/kustomize/api v0.0.1
	sigs.k8s.io/kustomize/v3 v3.3.1
	sigs.k8s.io/yaml v1.1.0
)

replace sigs.k8s.io/kustomize/api v0.0.1 => ../../../../api
