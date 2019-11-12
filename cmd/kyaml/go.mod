module sigs.k8s.io/kustomize/cmd/kyaml

go 1.13

require (
	github.com/go-errors/errors v1.0.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.4.0
	sigs.k8s.io/kustomize/kyaml v0.0.0
	sigs.k8s.io/kustomize/pseudo/k8s v0.0.0
)

replace (
	sigs.k8s.io/kustomize/kyaml v0.0.0 => ../../kyaml
	sigs.k8s.io/kustomize/pseudo/k8s v0.0.0 => ../../pseudo/k8s
)
