module sigs.k8s.io/kustomize/api

go 1.15

require (
	filippo.io/age v1.0.0-beta6.0.20210121110402-31e0d226807f
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/go-errors/errors v1.0.1
	github.com/go-openapi/spec v0.19.5
	github.com/golangci/golangci-lint v1.21.0
	github.com/google/go-cmp v0.3.0
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/hashicorp/go-multierror v1.1.0
	github.com/imdario/mergo v0.3.5
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.4.0
	github.com/yujunz/go-getter v1.5.1-lite.0.20201201013212-6d9c071adddf
	golang.org/x/crypto v0.0.0-20201208171446-5f87f3452ae9
	golang.org/x/term v0.0.0-20201117132131-f5c789dd3221
	golang.org/x/tools v0.0.0-20191119224855-298f0cb1881e
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.17.0
	k8s.io/kube-openapi v0.0.0-20191107075043-30be4d16710a
	sigs.k8s.io/kustomize/kyaml v0.10.6
	sigs.k8s.io/yaml v1.2.0
	sylr.dev/yaml/age/v3 v3.0.0-20210124165718-e183355958de
	sylr.dev/yaml/v3 v3.0.0-20210121142446-5fe289210a56
)

replace sigs.k8s.io/kustomize/kyaml => ../kyaml
