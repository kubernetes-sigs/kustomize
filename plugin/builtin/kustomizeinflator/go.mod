module sigs.k8s.io/kustomize/plugin/builtin/annotationstransformer

go 1.15

replace sigs.k8s.io/kustomize/kyaml => ../../../kyaml

require (
	github.com/sirupsen/logrus v1.8.0
	sigs.k8s.io/kustomize/api v0.8.4
	sigs.k8s.io/kustomize/kyaml v0.10.13
)
