module sigs.k8s.io/kustomize/cmd/pluginator/v2

go 1.16

require (
	github.com/rakyll/statik v0.1.7
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.5.1
	sigs.k8s.io/kustomize/api v0.8.8
	sigs.k8s.io/kustomize/kyaml v0.10.17
)

replace sigs.k8s.io/kustomize/api => ../../api

replace sigs.k8s.io/kustomize/kyaml => ../../kyaml

replace gopkg.in/yaml.v3 => gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
