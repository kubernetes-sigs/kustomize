module sigs.k8s.io/kustomize/kustomize/v3

go 1.13

require (
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	sigs.k8s.io/kustomize/api v0.1.1
	sigs.k8s.io/kustomize/forked/api v0.0.0
	sigs.k8s.io/kustomize/forked/apimachinery v0.0.0
	sigs.k8s.io/kustomize/forked/client-go v0.0.0
	sigs.k8s.io/yaml v1.1.0
)

replace (
	sigs.k8s.io/kustomize/api v0.1.1 => ../api
	sigs.k8s.io/kustomize/forked/api v0.0.0 => ../forked/api
	sigs.k8s.io/kustomize/forked/apimachinery v0.0.0 => ../forked/apimachinery
	sigs.k8s.io/kustomize/forked/client-go v0.0.0 => ../forked/client-go
)
