module sigs.k8s.io/kustomize/cmd/config

go 1.16

require (
	github.com/go-errors/errors v1.4.0
	github.com/kr/text v0.2.0 // indirect
	github.com/olekukonko/tablewriter v0.0.4
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	gopkg.in/inf.v0 v0.9.1
	k8s.io/kube-openapi v0.0.0-20210817084001-7fbd8d59e5b8
	sigs.k8s.io/kustomize/kyaml v0.11.1
)

replace sigs.k8s.io/kustomize/kyaml => ../../kyaml
