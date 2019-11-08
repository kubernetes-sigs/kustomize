module sigs.k8s.io/kustomize/cmd/kyaml

go 1.12

require (
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.4.0
	k8s.io/apimachinery v0.0.0-20191107105744-2c7f8d2b0fd8
	sigs.k8s.io/kustomize/kyaml v0.0.0
)

replace sigs.k8s.io/kustomize/kyaml v0.0.0 => ../../kyaml
