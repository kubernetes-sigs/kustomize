module sigs.k8s.io/kustomize/pluginator

go 1.13

require (
	cloud.google.com/go v0.38.0 // indirect
	github.com/Azure/go-autorest/autorest v0.9.0 // indirect
	github.com/emicklei/go-restful v2.9.6+incompatible // indirect
	github.com/golangci/golangci-lint v1.19.1 // indirect
	github.com/googleapis/gnostic v0.3.0 // indirect
	github.com/gophercloud/gophercloud v0.1.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20180305231024-9cad4c3443a7 // indirect
	github.com/imdario/mergo v0.3.5 // indirect
	github.com/monopole/mdrip v1.0.0 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/tools v0.0.0-20191107185733-c07e1c6ef61c // indirect
	google.golang.org/appengine v1.5.0 // indirect
	k8s.io/api v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/utils v0.0.0-20191030222137-2b95a09bc58d // indirect
	sigs.k8s.io/kustomize/api v0.1.1
)

replace (
	k8s.io/api v0.0.0 => ../forked/api
	k8s.io/apimachinery v0.0.0 => ../forked/apimachinery
	k8s.io/client-go v0.0.0 => ../forked/client-go
	sigs.k8s.io/kustomize/api v0.0.0 => ../api
)
