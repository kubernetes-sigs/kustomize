module sigs.k8s.io/kustomize/pseudo/k8s

go 1.13

require (
	github.com/Azure/go-autorest/autorest v0.9.2
	github.com/Azure/go-autorest/autorest/adal v0.8.0
	github.com/davecgh/go-spew v1.1.1
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c
	github.com/elazarl/goproxy v0.0.0-20191011121108-aa519ddbe484
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/gogo/protobuf v1.3.1
	github.com/golang/groupcache v0.0.0-20191027212112-611e8accdfc9
	github.com/golang/protobuf v1.2.0
	github.com/google/btree v1.0.0 // indirect
	github.com/google/go-cmp v0.3.1
	github.com/google/gofuzz v1.0.0
	github.com/google/uuid v1.1.1
	github.com/googleapis/gnostic v0.0.0-20170729233727-0c5108395e2d
	github.com/gophercloud/gophercloud v0.6.0
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/hashicorp/golang-lru v0.5.3
	github.com/imdario/mergo v0.3.5
	github.com/json-iterator/go v1.1.8
	github.com/modern-go/reflect2 v1.0.1
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f
	github.com/peterbourgon/diskv v2.0.1+incompatible
	github.com/pkg/errors v0.8.1 // indirect
	github.com/spf13/pflag v0.0.0-20170130214245-9ff6c6923cff
	github.com/stretchr/testify v1.3.0
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
	golang.org/x/net v0.0.0-20190620200207-3b0461eec859
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0
	gopkg.in/inf.v0 v0.9.1
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/klog v1.0.0
	k8s.io/kube-openapi v0.0.0-20191107075043-30be4d16710a
	k8s.io/utils v0.0.0-20191030222137-2b95a09bc58d
	sigs.k8s.io/yaml v1.1.0
)

replace golang.org/x/sys => golang.org/x/sys v0.0.0-20190813064441-fde4db37ae7a

replace golang.org/x/tools => golang.org/x/tools v0.0.0-20190821162956-65e3620a7ae7
