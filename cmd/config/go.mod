module sigs.k8s.io/kustomize/cmd/config

go 1.22.7

require (
	github.com/go-errors/errors v1.4.2
	github.com/spf13/cobra v1.8.0
	github.com/stretchr/testify v1.9.0
	gopkg.in/inf.v0 v0.9.1
	sigs.k8s.io/kustomize/kyaml v0.18.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/google/gnostic-models v0.6.9 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/kube-openapi v0.0.0-20241212222426-2c72e554b1e7 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)

replace sigs.k8s.io/kustomize/kyaml => ../../kyaml
