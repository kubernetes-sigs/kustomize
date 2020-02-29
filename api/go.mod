module sigs.k8s.io/kustomize/api

go 1.13

require (
	github.com/bgentry/go-netrc v0.0.0-20140422174119-9fd32a8b3d3d
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/go-openapi/spec v0.19.4
	github.com/golangci/golangci-lint v1.21.0
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/hashicorp/go-cleanhttp v0.5.1
	github.com/hashicorp/go-safetemp v1.0.0
	github.com/hashicorp/go-version v1.2.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/go-testing-interface v1.0.0
	github.com/pkg/errors v0.8.1
	github.com/ulikunitz/xz v0.5.7
	golang.org/x/tools v0.0.0-20191010075000-0337d82405ff
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.17.0
	k8s.io/kube-openapi v0.0.0-20191107075043-30be4d16710a
	sigs.k8s.io/yaml v1.1.0
)

replace sigs.k8s.io/kustomize/api/internal/getter/helper/url => ./helper/url
