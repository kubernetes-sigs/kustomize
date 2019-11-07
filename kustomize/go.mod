module sigs.k8s.io/kustomize/kustomize/v3

go 1.13

require (
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/client-go v0.0.0
	sigs.k8s.io/kustomize/api v0.0.0
	sigs.k8s.io/yaml v1.1.0
)

replace (
	k8s.io/api v0.0.0 => ../forked/api
	k8s.io/apimachinery v0.0.0 => ../forked/apimachinery
	k8s.io/client-go v0.0.0 => ../forked/client-go
	sigs.k8s.io/kustomize/api v0.0.0 => ../api
)
