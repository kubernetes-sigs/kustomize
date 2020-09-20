module sigs.k8s.io/kustomize/cmd/config

go 1.14

require (
	github.com/go-errors/errors v1.0.1
	github.com/go-openapi/spec v0.19.5
	github.com/olekukonko/tablewriter v0.0.4
	github.com/posener/complete/v2 v2.0.1-alpha.12
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	k8s.io/apimachinery v0.17.3
	k8s.io/cli-runtime v0.17.3
	k8s.io/client-go v0.17.3
	k8s.io/kubectl v0.0.0-20191219154910-1528d4eea6dd
	sigs.k8s.io/cli-utils v0.20.2
	sigs.k8s.io/kustomize/kyaml v0.8.1
)

replace sigs.k8s.io/kustomize/kyaml v0.8.1 => ../../kyaml
