module sigs.k8s.io/kustomize/hack/crawl

go 1.13

require (
	github.com/elastic/go-elasticsearch/v6 v6.8.2
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/googleapis/gnostic v0.3.0 // indirect
	github.com/gorilla/mux v1.7.3
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/json-iterator/go v1.1.8 // indirect
	github.com/rs/cors v1.7.0
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/kube-openapi v0.0.0-20191107075043-30be4d16710a // indirect
	sigs.k8s.io/kustomize/api v0.1.1
	sigs.k8s.io/yaml v1.1.0
)

replace sigs.k8s.io/kustomize/api v0.0.0 => ../../api
