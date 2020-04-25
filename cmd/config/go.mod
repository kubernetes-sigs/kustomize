module sigs.k8s.io/kustomize/cmd/config

go 1.13

require (
	github.com/go-errors/errors v1.0.1
	github.com/olekukonko/tablewriter v0.0.4
	github.com/posener/complete/v2 v2.0.1-alpha.12
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	k8s.io/apimachinery v0.17.0
	sigs.k8s.io/kustomize/kyaml v0.1.4
)

replace sigs.k8s.io/kustomize/kyaml v0.1.4 => ../../kyaml
