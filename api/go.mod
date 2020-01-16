module sigs.k8s.io/kustomize/api

go 1.13

require (
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/go-openapi/spec v0.19.4
	github.com/golangci/golangci-lint v1.21.0
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/hairyhenderson/gomplate/v3 v3.6.0
	github.com/imdario/mergo v0.3.5
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.4.0
	golang.org/x/tools v0.0.0-20191010075000-0337d82405ff
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20191107075043-30be4d16710a
	sigs.k8s.io/yaml v1.1.0
)

replace k8s.io/client-go => k8s.io/client-go v0.17.0
