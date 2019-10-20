module sigs.k8s.io/kustomize/internal/tools

go 1.13

require (
	github.com/elastic/go-elasticsearch/v6 v6.8.2
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/gorilla/mux v1.7.3
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/rs/cors v1.7.0
	sigs.k8s.io/kustomize/v3 v3.3.1
	sigs.k8s.io/yaml v1.1.0
)

replace sigs.k8s.io/kustomize/v3 v3.3.1 => ../../
