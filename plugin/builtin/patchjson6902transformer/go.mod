module sigs.k8s.io/kustomize/plugin/builtin/patchjson6902transformer

go 1.16

require (
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/evanphx/json-patch v4.11.0+incompatible
	github.com/pkg/errors v0.9.1
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4 // indirect
	sigs.k8s.io/kustomize/api v0.8.9
	sigs.k8s.io/kustomize/kyaml v0.13.6
	sigs.k8s.io/yaml v1.2.0
)

replace sigs.k8s.io/kustomize/api => ../../../api

replace sigs.k8s.io/kustomize/kyaml => ../../../kyaml
