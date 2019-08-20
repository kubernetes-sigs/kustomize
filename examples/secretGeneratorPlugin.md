[ConfigMaps]: https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#configmap-v1-core
[ELF]: https://en.wikipedia.org/wiki/Executable_and_Linkable_Format
[Go plugin]: https://golang.org/pkg/plugin
[Secrets]: https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#secret-v1-core
[base64]: https://tools.ietf.org/html/rfc4648#section-4
[configuration directory]: https://wiki.archlinux.org/index.php/XDG_Base_Directory#Specification
[grpc]: https://grpc.io
[tag]: https://github.com/kubernetes-sigs/kustomize/releases
[v2.0.3]: https://github.com/kubernetes-sigs/kustomize/releases/tag/v2.0.3
[`exec.Command`]: https://golang.org/pkg/os/exec/#Command

# Generating Secrets

## What's a Secret?

Kubernetes [ConfigMaps] and [Secrets] are both
key:value maps, but the latter is intended to
signal that its values have a sensitive nature -
e.g. pass phrases or ssh keys.

Kubernetes developers work in various ways to hide
the information in a Secret more carefully than
the information held by ConfigMaps, Deployments,
etc.

## Make a place to work

<!-- @establishBase @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)
```

## Secret values from local files

kustomize has three different (builtin) ways to
generate a secret from local files:

 * get them from so-called _env_ files (`NAME=VALUE`, one per line),
 * consume the entire contents of a file to make one secret value,
 * get literal values from the kustomization file itself.

Here's an example combining all three methods:

Make an env file with some short secrets:

<!-- @makeEnvFile @testAgainstLatestRelease -->
```
cat <<'EOF' >$DEMO_HOME/foo.env
ROUTER_PASSWORD=admin
DB_PASSWORD=iloveyou
EOF
```

Make a text file with a long secret:

<!-- @makeLongSecretFile @testAgainstLatestRelease -->
```
cat <<'EOF' >$DEMO_HOME/longsecret.txt
Lorem ipsum dolor sit amet,
consectetur adipiscing elit,
sed do eiusmod tempor incididunt
ut labore et dolore magna aliqua.
EOF
```

And make a kustomization file referring to the
above and _additionally_ defining some literal KV
pairs in line:

<!-- @makeKustomization1 @testAgainstLatestRelease -->
```
cat <<'EOF' >$DEMO_HOME/kustomization.yaml
secretGenerator:
- name: mysecrets
  envs:
  - foo.env
  files:
  - longsecret.txt
  literals:
  - FRUIT=apple
  - VEGETABLE=carrot
EOF
```

Now generate the Secret:

<!-- @build1 @testAgainstLatestRelease -->
```
result=$(kustomize build $DEMO_HOME)
echo "$result"
# Spot check the result:
test 1 == $(echo "$result" | grep -c "FRUIT: YXBwbGU=")
```

This emits something like

> ```
> apiVersion: v1
> kind: Secret
> metadata:
>   name: mysecrets-hfb5df789h
> type: Opaque
> data:
>   FRUIT: YXBwbGU=
>   VEGETABLE: Y2Fycm90
>   ROUTER_PASSWORD: YWRtaW4=
>   DB_PASSWORD: aWxvdmV5b3U=
>   longsecret.txt: TG9yZW0gaXBzdW0gZG9sb3Igc2l0I... (elided)
> ```

The name of the resource is the prefix `mysecrets`
(as specfied in the kustomization file), followed
by a hash of its contents.

Use your favorite base64 decoder to confirm the raw
versions of any of these values.

The problem that these three approaches share is
that the purported secrets must live on disk.

This adds additional security questions - who can
see the files, who installs them, who deletes
them, etc.


## Secret values from anywhere

A general alternative is to enshrine secret
value generation in a [plugin](../docs/plugins).

The values can then come in via, say, an
authenticated and authorized RPC to a password
vault service.

[sgp]: ../plugin/someteam.example.com/v1/secretsfromdatabase

Here's a [secret generator plugin][sgp]
that pretends to pull the values of a map
from a database.


Download it

<!-- @copyPlugin @testAgainstLatestRelease -->
```
repo=https://raw.githubusercontent.com/kubernetes-sigs/kustomize
pPath=plugin/someteam.example.com/v1/secretsfromdatabase
dir=$DEMO_HOME/kustomize/$pPath

mkdir -p $dir

curl -s -o $dir/SecretsFromDatabase.go \
  ${repo}/master/$pPath/SecretsFromDatabase.go
```

Compile it

<!-- @compilePlugin @xtest -->
```
go build -buildmode plugin \
  -o $dir/SecretsFromDatabase.so \
  $dir/SecretsFromDatabase.go
```


Create a configuration file for it:

<!-- @makeConfiguration @testAgainstLatestRelease -->
```
cat <<'EOF' >$DEMO_HOME/secretFromDb.yaml
apiVersion: someteam.example.com/v1
kind: SecretsFromDatabase
metadata:
  name: mySecretGenerator
name: forbiddenValues
namespace: production
keys:
- ROCKET
- VEGETABLE
EOF
```

Create a new kustomization file
referencing this plugin:

<!-- @makeKustomization2 @testAgainstLatestRelease -->
```
cat <<'EOF' >$DEMO_HOME/kustomization.yaml
generators:
- secretFromDb.yaml
EOF
```

Finally, generate the secret, setting
`XDG_CONFIG_HOME` so that the plugin
can be found under `$DEMO_HOME`:

<!-- @build2 @xtest -->
```
result=$( \
  XDG_CONFIG_HOME=$DEMO_HOME \
  kustomize build --enable_alpha_plugins $DEMO_HOME )
echo "$result"
# Spot check the result:
test 1 == $(echo "$result" | grep -c "FRUIT: YXBwbGU=")
```

This should emit something like:

> ```
> apiVersion: v1
> kind: Secret
> metadata:
>   name: mysecrets-bdt27dbkd6
> type: Opaque
> data:
>  FRUIT: YXBwbGU=
>  VEGETABLE: Y2Fycm90
> ```

i.e. a subset of the same values as above.
