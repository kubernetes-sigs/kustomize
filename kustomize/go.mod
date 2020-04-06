module sigs.k8s.io/kustomize/kustomize/v3

go 1.13

require (
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.4.0
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v11.0.0+incompatible
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/kustomize/api v0.3.2
	sigs.k8s.io/kustomize/cmd/config v0.0.5
	sigs.k8s.io/kustomize/cmd/kubectl v0.0.3
	sigs.k8s.io/kustomize/kstatus v0.0.1
	sigs.k8s.io/kustomize/kyaml v0.0.6
	sigs.k8s.io/yaml v1.1.0
)

exclude (
	github.com/Azure/go-autorest v12.0.0+incompatible
	github.com/russross/blackfriday v2.0.0+incompatible
	sigs.k8s.io/kustomize/api v0.2.0
)

replace (
	github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.0
	k8s.io/client-go => k8s.io/client-go v0.17.0
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20191219154910-1528d4eea6dd
	sigs.k8s.io/kustomize/api => ../api
	sigs.k8s.io/kustomize/cmd/kubectl v0.0.3 => ../cmd/kubectl
	sigs.k8s.io/kustomize/kstatus v0.0.1 => ../kstatus
)
