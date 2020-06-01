module sigs.k8s.io/kustomize/cmd/kubectl

go 1.13

require (
	github.com/go-errors/errors v1.0.1
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.4.0
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/cli-runtime v0.17.0
	k8s.io/client-go v0.17.0
	k8s.io/component-base v0.17.0 // indirect
	k8s.io/kubectl v0.0.0-20191219154910-1528d4eea6dd
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/kustomize/kstatus v0.0.2
	sigs.k8s.io/kustomize/kyaml v0.1.11
)
