module sigs.k8s.io/kustomize/kustomize/v3

go 1.13

require (
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	k8s.io/client-go v0.17.0
	sigs.k8s.io/kustomize/api v0.3.1
	sigs.k8s.io/kustomize/cmd/config v0.0.2
	sigs.k8s.io/kustomize/cmd/kubectl v0.0.2
	sigs.k8s.io/kustomize/kyaml v0.0.2
	sigs.k8s.io/yaml v1.1.0
)

exclude (
	github.com/russross/blackfriday v2.0.0+incompatible
	sigs.k8s.io/kustomize/api v0.2.0
)

replace sigs.k8s.io/kustomize/api v0.3.1 => ../api
