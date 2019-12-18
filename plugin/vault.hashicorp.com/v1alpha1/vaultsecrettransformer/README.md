## Introduction
This is an attempt to demonstrate the power of Kustomize plugin mechanism. [Written earlier documentation](/docs/plugins/goPluginGuidedExample.md) contains concise instructions how to build a very simple plugin.

The plugin introduced here is a bit more complicated example. It carries out one useful function of pulling secrets from [Vault](https://www.vaultproject.io/) and populating Kustomize resource scope with them. It additionally involves [CRDs](/plugin/vault.hashicorp.com/v1alpha1/vaultsecrettransformer/test/crd-vaultsecret.yaml) with a simple DSL where [a secret specification](/plugin/vault.hashicorp.com/v1alpha1/vaultsecrettransformer/test/vaultsecret.yaml) is defined.
 
The plugin needs to be invoked only at the most top level of inclusion if you operate with [a multi-layer layout](https://kubectl.docs.kubernetes.io/pages/app_composition_and_deployment/structure_directories.html). Once it's launched, it looks through the scope, detects custom resources, pulls them from Vault, and creates [standard Secret resources](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#secret-v1-core).
 
There's a capability to manage *Opaque* and *dockerconfigjson* secret types.

## Usage
To see how the plugin can be used, have a look at [the Integration Tests section](#Integration Tests)

## Build
To be able to use this plugin, Kustomize binary with matched dependencies has to be built. You can try though a prebuild binary by downloading it from [the release page](https://github.com/kubernetes-sigs/kustomize/releases) but you need to make sure you are using the release the plugin has been developed for (v3.5.3 at the time of writing).

Since you need to compile the plugin anyway, it makes sense to build both binaries Kustomize and VaultSecretTransformer.so: 

```shell script
PWD=$(pwd)
git clone git@github.com:kubernetes-sigs/kustomize.git

cd "$PWD/kustomize/kustomize"
go build

cd "$PWD/kustomize/plugin/vault.hashicorp.com/v1alpha1/vaultsecrettransformer"
go build -buildmode plugin -o VaultSecretTransformer.so
```

## Tests
There are two types of tests ensuring the plugin's functionality. Unit tests implemented in Go and Integration tests using standard Kustomize invocation mechanism (YAMLs and BASH).

### Unit Tests
[Unit tests](VaultSecretTransformer_test.go) cover part of the code responsible for generating valid secret resource object. You might find it useful if decide to bring in some logic how secrets are handled.

You can run unit tests using Go CLI:
```shell script
cd "$PWD/kustomize/plugin/vault.hashicorp.com/v1alpha1/vaultsecrettransformer"
go test
```

### Integration Tests
Integration test involves interaction with Vault and requires Vault environment variable to be provided before running the test:

```shell script
export VAULT_ADDR=https://vault.yourdomain.com
export VAULT_TOKEN=<YOUR_TOKEN>
```

Kustomize operates KDG Base Directory specification and requires the *XDG_CONFIG_HOME* environment variable to be defined.
```shell script
export XDG_CONFIG_HOME=$PWD
```

The test represents the usual way of running Kustomize with plugins. It sets the main [kustomization.yaml](test/kustomization.yaml) file with the declaration of resources needed for pulling a secret. To run the test, type the following:
```shell script
cd "$PWD/kustomize/plugin/vault.hashicorp.com/v1alpha1/vaultsecrettransformer/test"
$PWD/kustomize/kustomize/kustomize build --enable_alpha_plugins . -o variant.yaml
```
