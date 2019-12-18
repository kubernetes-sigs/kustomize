module sigs.k8s.io/kustomize/plugin/vault.hashicorp.com/v1alpha/vaultsecrettransformer

go 1.13

require (
	github.com/gogo/protobuf v1.2.2-0.20190723190241-65acae22fc9d
	github.com/hashicorp/vault/api v1.0.4
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	sigs.k8s.io/kustomize/api v0.3.1
	sigs.k8s.io/yaml v1.1.0
)

exclude sigs.k8s.io/kustomize/api v0.2.0
