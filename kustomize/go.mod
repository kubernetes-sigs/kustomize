module sigs.k8s.io/kustomize/kustomize/v5

go 1.24.0

require (
	github.com/google/go-cmp v0.7.0
	github.com/spf13/cobra v1.9.1
	github.com/spf13/pflag v1.0.6
	github.com/stretchr/testify v1.10.0
	golang.org/x/text v0.28.0
	sigs.k8s.io/kustomize/api v0.21.0
	sigs.k8s.io/kustomize/cmd/config v0.21.0
	sigs.k8s.io/kustomize/kyaml v0.21.0
	sigs.k8s.io/yaml v1.5.0
	github.com/opencontainers/image-spec v1.1.1
	oras.land/oras-go/v2 v2.6.0
)

require (
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-errors/errors v1.4.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/google/gnostic-models v0.6.9 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sergi/go-diff v1.4.0 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	go.yaml.in/yaml/v3 v3.0.3 // indirect
	golang.org/x/sys v0.35.0 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/kube-openapi v0.0.0-20241212222426-2c72e554b1e7 // indirect
)

replace sigs.k8s.io/kustomize/api => ../api

replace sigs.k8s.io/kustomize/cmd/config => ../cmd/config

replace sigs.k8s.io/kustomize/kyaml => ../kyaml
