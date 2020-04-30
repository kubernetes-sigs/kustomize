module sigs.k8s.io/kustomize/cmd/config

go 1.13

require (
	github.com/coreos/go-etcd v2.0.0+incompatible // indirect
	github.com/cpuguy83/go-md2man v1.0.10 // indirect
	github.com/go-errors/errors v1.0.1
	github.com/olekukonko/tablewriter v0.0.4
	github.com/posener/complete/v2 v2.0.1-alpha.12
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	github.com/ugorji/go/codec v0.0.0-20181204163529-d75b2dcb6bc8 // indirect
	k8s.io/apimachinery v0.17.0
	sigs.k8s.io/kustomize/kyaml v0.0.0
)

replace sigs.k8s.io/kustomize/kyaml v0.0.0 => ../../kyaml
