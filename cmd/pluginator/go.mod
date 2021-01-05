module sigs.k8s.io/kustomize/cmd/pluginator/v2

go 1.15

require (
	github.com/rakyll/statik v0.1.7
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.4.0
	sigs.k8s.io/kustomize/api v0.7.1
	sigs.k8s.io/kustomize/kyaml v0.10.5
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/kustomize/api => ../../api

replace sigs.k8s.io/kustomize/kyaml => ../../kyaml
