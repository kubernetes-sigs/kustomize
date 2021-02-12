module sigs.k8s.io/kustomize/api

go 1.15

require (
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/go-errors/errors v1.0.1
	github.com/go-openapi/spec v0.19.5
	github.com/golangci/golangci-lint v1.21.0
	github.com/google/go-cmp v0.3.0
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/imdario/mergo v0.3.5
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.4.0
	golang.org/x/tools v0.0.0-20191119224855-298f0cb1881e
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
	sigs.k8s.io/kustomize/kyaml v0.10.9
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/kustomize/kyaml => ../kyaml
