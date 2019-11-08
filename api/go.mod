module sigs.k8s.io/kustomize/api

go 1.13

require (
	github.com/emicklei/go-restful v2.9.6+incompatible // indirect
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/go-openapi/spec v0.19.4
	github.com/golangci/golangci-lint v1.19.1
	github.com/googleapis/gnostic v0.3.0 // indirect
	github.com/monopole/mdrip v1.0.0
	github.com/pkg/errors v0.8.1
	golang.org/x/tools v0.0.0-20190912215617-3720d1ec3678
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/kube-openapi v0.0.0-20190918143330-0270cf2f1c1d
	sigs.k8s.io/kustomize/forked/api v0.0.0
	sigs.k8s.io/kustomize/forked/apimachinery v0.0.0
	sigs.k8s.io/kustomize/forked/client-go v0.0.0
	sigs.k8s.io/yaml v1.1.0
)

replace (
	sigs.k8s.io/kustomize/forked/api v0.0.0 => ../forked/api
	sigs.k8s.io/kustomize/forked/apimachinery v0.0.0 => ../forked/apimachinery
	sigs.k8s.io/kustomize/forked/client-go v0.0.0 => ../forked/client-go

)
