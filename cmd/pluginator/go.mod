module sigs.k8s.io/kustomize/cmd/pluginator/v2

go 1.20

require (
	github.com/spf13/cobra v1.7.0
	github.com/stretchr/testify v1.8.1
	sigs.k8s.io/kustomize/api v0.15.0
	sigs.k8s.io/kustomize/kyaml v0.15.0
)

replace sigs.k8s.io/kustomize/api => ../../api

replace sigs.k8s.io/kustomize/kyaml => ../../kyaml
