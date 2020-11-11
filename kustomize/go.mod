module sigs.k8s.io/kustomize/kustomize/v3

go 1.14

require (
	github.com/google/go-cmp v0.5.2
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	k8s.io/client-go v0.18.10
	sigs.k8s.io/kustomize/api v0.6.5
	sigs.k8s.io/kustomize/cmd/config v0.8.5
	sigs.k8s.io/kustomize/kyaml v0.9.4
	sigs.k8s.io/yaml v1.2.0
)

exclude (
	github.com/russross/blackfriday v2.0.0+incompatible
	sigs.k8s.io/kustomize/api v0.2.0
	sigs.k8s.io/kustomize/cmd/config v0.2.0
)
