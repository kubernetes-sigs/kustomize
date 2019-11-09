module sigs.k8s.io/kustomize/cmd/kyaml

go 1.12

require (
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.4.0
	sigs.k8s.io/kustomize/kyaml v0.0.0
	sigs.k8s.io/kustomize/pseudo/k8s v0.0.0-20191109010559-74255f6badd9
)

replace sigs.k8s.io/kustomize/kyaml v0.0.0 => ../../kyaml
