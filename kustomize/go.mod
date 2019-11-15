module sigs.k8s.io/kustomize/kustomize/v3

go 1.13

require (
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	sigs.k8s.io/kustomize/api v0.2.0
	sigs.k8s.io/kustomize/cmd/cfg v0.0.0
	sigs.k8s.io/yaml v1.1.0
)

replace (
	sigs.k8s.io/kustomize/cmd/cfg v0.0.0 => ../cmd/cfg
	sigs.k8s.io/kustomize/kyaml v0.0.0 => ../kyaml
	sigs.k8s.io/kustomize/pseudo/k8s v0.0.0 => ../pseudo/k8s
)
