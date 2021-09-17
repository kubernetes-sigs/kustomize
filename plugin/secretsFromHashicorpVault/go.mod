module github.com/benmorehouse/kustomize/plugin/secretsFromHashicorpVault

go 1.16

require (
	github.com/hashicorp/vault v1.8.2
	github.com/hashicorp/vault/api v1.1.2-0.20210713235431-1fc8af4c041f
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.16.0
	sigs.k8s.io/kustomize/api v0.9.0
	sigs.k8s.io/yaml v1.2.0
)
