module sigs.k8s.io/kustomize/cmd/pluginator/v2

go 1.16

require (
	github.com/rakyll/statik v0.1.7
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.5.1
	sigs.k8s.io/kustomize/api v0.8.10
	sigs.k8s.io/kustomize/kyaml v0.10.20
)

replace gopkg.in/yaml.v3 => github.com/natasha41575/yaml v0.0.0-20210623011331-77123dad73ab

replace sigs.k8s.io/kustomize/api => ../../api

replace sigs.k8s.io/kustomize/kyaml => ../../kyaml
