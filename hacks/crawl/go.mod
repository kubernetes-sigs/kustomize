module sigs.k8s.io/kustomize/internal/tools

go 1.13

require (
	github.com/elastic/go-elasticsearch/v6 v6.8.2
	github.com/emicklei/go-restful v2.9.6+incompatible // indirect
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/gofuzz v1.0.0 // indirect
	github.com/googleapis/gnostic v0.3.0 // indirect
	github.com/gorilla/mux v1.7.3
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/json-iterator/go v1.1.6 // indirect
	github.com/mailru/easyjson v0.0.0-20190620125010-da37f6c1e481 // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/rs/cors v1.7.0
	k8s.io/klog v0.3.3 // indirect
	sigs.k8s.io/kustomize/v3 v3.3.1
	sigs.k8s.io/yaml v1.1.0
)

replace sigs.k8s.io/kustomize/v3 v3.3.1 => ../../
