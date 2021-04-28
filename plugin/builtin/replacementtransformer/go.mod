module sigs.k8s.io/kustomize/plugin/builtin/replacementtransformer

go 1.16

require (
	sigs.k8s.io/kustomize/api v0.0.0
	sigs.k8s.io/kustomize/kyaml v0.10.17
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/kustomize/kyaml => ./../../../kyaml

replace sigs.k8s.io/kustomize/api => ./../../../api

replace gopkg.in/yaml.v3 => gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
