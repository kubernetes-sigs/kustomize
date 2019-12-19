module sigs.k8s.io/kustomize/cmd/kubectl

go 1.13

require (
	github.com/spf13/cobra v0.0.5
	k8s.io/cli-runtime v0.17.0
	k8s.io/client-go v0.17.0
	k8s.io/component-base v0.17.0 // indirect
	k8s.io/kubectl v0.0.0-20191219154910-1528d4eea6dd
	sigs.k8s.io/kustomize/kyaml v0.0.0
)

replace sigs.k8s.io/kustomize/kyaml v0.0.0 => ../../kyaml
