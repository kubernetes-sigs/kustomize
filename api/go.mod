module sigs.k8s.io/kustomize/api

go 1.14

require (
	github.com/Masterminds/semver/v3 v3.1.0
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/go-openapi/spec v0.19.5
	github.com/gofrs/flock v0.7.1
	github.com/golangci/golangci-lint v1.21.0
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/google/uuid v1.1.1
	github.com/hairyhenderson/gomplate/v3 v3.6.0
	github.com/hashicorp/go-getter v1.4.1
	github.com/imdario/mergo v0.3.8
	github.com/kr/text v0.2.0 // indirect
	github.com/mholt/archiver/v3 v3.3.0
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pkg/errors v0.9.1
	github.com/sosedoff/gitkit v0.2.0
	github.com/stretchr/testify v1.4.0
	github.com/yujunz/go-getter v1.4.1-lite
	golang.org/x/crypto v0.0.0-20200302210943-78000ba7a073 // indirect
	golang.org/x/net v0.0.0-20200301022130-244492dfa37a // indirect
	golang.org/x/sys v0.0.0-20200302150141-5c8b2ff67527 // indirect
	golang.org/x/tools v0.0.0-20191119224855-298f0cb1881e
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200121175148-a6ecf24a6d71
	helm.sh/helm/v3 v3.1.2
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20191107075043-30be4d16710a
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/kustomize/kyaml v0.8.1
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.0
	k8s.io/client-go => k8s.io/client-go v0.17.0
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20191219154910-1528d4eea6dd
)

exclude github.com/Azure/go-autorest v12.0.0+incompatible
