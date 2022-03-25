module sigs.k8s.io/kustomize/cmd/pluginator/v2

go 1.16

require (
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/rakyll/statik v0.1.7
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4 // indirect
	sigs.k8s.io/kustomize/api v0.11.3
	sigs.k8s.io/kustomize/kyaml v0.13.4
)

replace sigs.k8s.io/kustomize/api => ../../api

replace sigs.k8s.io/kustomize/kyaml => ../../kyaml
