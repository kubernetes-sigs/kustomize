module sigs.k8s.io/kustomize/plugin/someteam.example.com/v1/sedtransformer

go 1.16

require (
	sigs.k8s.io/kustomize/api v0.8.9
	sigs.k8s.io/kustomize/kyaml v0.13.0 // indirect
)

replace sigs.k8s.io/kustomize/api => ../../../../api

replace sigs.k8s.io/kustomize/kyaml => ../../../../kyaml
