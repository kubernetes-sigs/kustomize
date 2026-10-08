module sigs.k8s.io/kustomize/plugin/builtin/patchtransformer

go 1.27.0

require (
	github.com/stretchr/testify v1.12.1
	gopkg.in/evanphx/json-patch.v4 v4.13.0
	sigs.k8s.io/kustomize/api v0.21.2
	sigs.k8s.io/kustomize/kyaml v0.21.3
	sigs.k8s.io/yaml v1.6.0
)

require (
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/go-errors/errors v1.5.1 // indirect
	github.com/go-openapi/jsonpointer v1.0.0 // indirect
	github.com/go-openapi/jsonreference v1.0.0 // indirect
	github.com/google/gnostic-models v0.7.1 // indirect
	github.com/monochromegane/go-gitignore v0.0.0-20200626010858-205db1a8cc00 // indirect
	github.com/xlab/treeprint v1.2.0 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/sys v0.48.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	k8s.io/kube-openapi v0.0.0-20261001230523-97fa35140926 // indirect
)

replace sigs.k8s.io/kustomize/api => ../../../api
