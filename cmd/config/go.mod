module sigs.k8s.io/kustomize/cmd/config

go 1.13

require (
	github.com/go-errors/errors v1.0.1
	github.com/posener/complete/v2 v2.0.1-alpha.12
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	k8s.io/apimachinery v0.17.0
	sigs.k8s.io/kustomize/kyaml v0.0.0
)

replace sigs.k8s.io/kustomize/kyaml v0.0.0 => ../../kyaml
