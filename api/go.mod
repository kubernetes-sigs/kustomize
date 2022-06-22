module sigs.k8s.io/kustomize/api

go 1.16

require (
	github.com/a8m/envsubst v1.3.0
	github.com/evanphx/json-patch v4.11.0+incompatible
	github.com/go-errors/errors v1.0.1
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/imdario/mergo v0.3.5
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/kube-openapi v0.0.0-20220401212409-b28bf2818661
	sigs.k8s.io/kustomize/kyaml v0.13.7
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/kustomize/kyaml => ../kyaml
