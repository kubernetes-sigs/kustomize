module sigs.k8s.io/kustomize/kustomize/v3

go 1.13

require (
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.17.0
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/kustomize/api v0.4.1
	sigs.k8s.io/kustomize/cmd/config v0.1.11
	sigs.k8s.io/kustomize/cmd/kubectl v0.1.0
	sigs.k8s.io/kustomize/kstatus v0.0.2
	sigs.k8s.io/kustomize/kyaml v0.1.11
	sigs.k8s.io/yaml v1.2.0
)

exclude (
	github.com/russross/blackfriday v2.0.0+incompatible
	sigs.k8s.io/kustomize/api v0.2.0
)
