module sigs.k8s.io/kustomize/api

go 1.13

require (
	github.com/Masterminds/semver/v3 v3.1.0
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/go-git/go-git/v5 v5.0.0
	github.com/go-openapi/spec v0.19.4
	github.com/gofrs/flock v0.7.1
	github.com/golangci/golangci-lint v1.21.0
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/google/uuid v1.1.1
	github.com/hairyhenderson/gomplate/v3 v3.6.0
	github.com/imdario/mergo v0.3.8
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.4.0
	github.com/yujunz/go-getter v1.4.1-lite
	golang.org/x/tools v0.0.0-20191010075000-0337d82405ff
	gopkg.in/yaml.v2 v2.2.4
	helm.sh/helm/v3 v3.1.2
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20191107075043-30be4d16710a
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/yaml v1.1.0
)

replace (
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.0
	k8s.io/client-go => k8s.io/client-go v0.17.0
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20191219154910-1528d4eea6dd
)

exclude github.com/Azure/go-autorest v12.0.0+incompatible
