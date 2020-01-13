module sigs.k8s.io/kustomize/cmd/resource

go 1.12

require (
	github.com/acarl005/stripansi v0.0.0-20180116102854-5a71ef0e047d
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.4.0
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.17.0
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/kustomize/kstatus v0.0.0
	sigs.k8s.io/kustomize/kyaml v0.0.0
)

replace (
	sigs.k8s.io/kustomize/kstatus v0.0.0 => ../../kstatus
	sigs.k8s.io/kustomize/kyaml v0.0.0 => ../../kyaml
)
