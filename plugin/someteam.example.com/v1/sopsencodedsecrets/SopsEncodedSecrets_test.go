// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package main_test

import (
	"testing"

	"sigs.k8s.io/kustomize/pkg/kusttest"
	"sigs.k8s.io/kustomize/plugin"
)

func TestSopsEncodedSecretsPlugin(t *testing.T) {
	tc := plugin.NewEnvForTest(t).Set()
	defer tc.Reset()

	tc.BuildGoPlugin(
		"someteam.example.com", "v1", "SopsEncodedSecrets")

	th := kusttest_test.NewKustTestPluginHarness(t, "/app")

	/*

	   # Writing a portable test for sops is problematic,
	   # because sops decoding assumes access to a local
	   # private key in some form, and these test need
	   # to run anywhere, and they don't use a real file
	   # system.  Need to revisit this;
	   # maybe we can stick the private key in an ENV var?
	   # And use GPG instead of gcp_kms?

	   # To try this plugin by itself with real data
	   # in Google cloud kms, do the following:

	   gcloud kms keyrings create sops --location global
	   gcloud kms keys create sops-key --location global \
	       --keyring sops --purpose encryption
	   gcloud kms keys list --location global --keyring sops

	   project=$(\
	       gcloud kms keys list --location global --keyring sops |\
	       grep GOOGLE | cut -d" " -f1)
	   echo $project

	   go get -u go.mozilla.org/sops/cmd/sops

	   cat <<'EOF' >/tmp/sec_clear.yaml
	   VEGETABLE: carrot
	   ROCKET: saturn-v
	   FRUIT: apple
	   CAR: dymaxion
	   EOF

	   # Put the output of the following command into
	   # the encodedFileContent constant below:
	   sops --encrypt --gcp-kms $project /tmp/sec_clear.yaml

	*/
	const encodedFileContent = `
VEGETABLE: ENC[AES256_GCM,data:9mKo4gCm,iv:nkhvWPDbMkDeLXAhTxQOsCaz3ACAx4VS9CLR3tGe5zI=,tag:KIY4z/eE3DFnKHbHHB0ytQ==,type:str]
ROCKET: ENC[AES256_GCM,data:6C7vnZYkh+Q=,iv:66/EAqulH7OtMMvSyMZSL5ZbktEm4Yj5S7g/Zb+XgUk=,tag:yEaxZs57fKn7Uebk+ouDDw==,type:str]
FRUIT: ENC[AES256_GCM,data:2a/KQxA=,iv:7GmWqc6uA6h539DQVpGq8m0WZLAUi9jzZ6iQAnDEY0s=,tag:ItvY4ziCEW3yNLo/YKMxnw==,type:str]
CAR: ENC[AES256_GCM,data:SZFq30w5NZE=,iv:paZ+ghcYoIVIvuGvKP6K6+K7hIgS/l3KgoBxjzjIBHs=,tag:iNL2kvYMppDRXuybmsUFRw==,type:str]
sops:
    kms: []
    gcp_kms:
    -   resource_id: projects/__ELIDED_FOR_KUSTOMIZE_TEST__/locations/global/keyRings/sops/cryptoKeys/sops-key
        created_at: '2019-06-19T22:32:52Z'
        enc: __ELIDED_FOR_KUSTOMIZE_TEST__=
    azure_kv: []
    lastmodified: '2019-06-19T22:32:52Z'
    mac: ENC[AES256_GCM,data:__ELIDED_FOR_KUSTOMIZE_TEST__:str]
    pgp: []
    unencrypted_suffix: _unencrypted
    version: 3.3.1
`

	th.WriteF("/app/mySecrets.yaml", encodedFileContent)

	m := th.LoadAndRunGenerator(`
apiVersion: someteam.example.com/v1
kind: SopsEncodedSecrets
metadata:
  name: mySecretGenerator
name: forbiddenValues
namespace: production
file: mySecrets.yaml
keys:
- ROCKET
- CAR
`)
	th.AssertActualEqualsExpected(m, `
apiVersion: v1
data:
  CAR: ZHltYXhpb24=
  ROCKET: c2F0dXJuLXY=
kind: Secret
metadata:
  name: forbiddenValues
  namespace: production
type: Opaque
`)
}
