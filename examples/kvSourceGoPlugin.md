[Secrets]: https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#secret-v1-core
[ConfigMaps]: https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.13/#configmap-v1-core
[base64]: https://tools.ietf.org/html/rfc4648#section-4
[Go plugin]: https://golang.org/pkg/plugin
[v2.0.3]: https://github.com/kubernetes-sigs/kustomize/releases/tag/v2.0.3

# Generating Secrets

## What's a Secret?

Kubernetes [ConfigMaps] and [Secrets] are both
key:value (KV) maps, but the latter is intended to
signal that its values have a sensitive nature -
e.g. ssh keys or passwords.

Kubernetes assumes that the values in a Secret are
[base64] encoded, and decodes them before actual
use (as, say, the argument to a container
command).  The user that creates the Secret must
base64 encode the data, or use a tool that does it
for them.  This encoding doesn't protect the
secret from anything other than an
over-the-shoulder glance.

Protecting the actual secrecy of a Secret value is
up to the cluster operator. They must lock down
the cluster (and its `etcd` data store) as tightly
as desired, and likewise protect the bytes that
feed into the cluster to ultimately become the
content of a Secret value.

## Make a place to work

<!-- @establishBase @test -->
```
DEMO_HOME=$(mktemp -d)
```

## Secret values from local files

kustomize has three different ways to generate a secret
from local files:

 * get them from so-called _env_ files (`NAME=VALUE`, one per line),
 * consume the entire contents of a file to make one secret value,
 * get literal values from the kustomization file itself.
 
Here's an example combining all three methods:

Make an env file with some short secrets:

<!-- @makeEnvFile @test -->
```
cat <<'EOF' >$DEMO_HOME/foo.env
ROUTER_PASSWORD=admin
DB_PASSWORD=iloveyou
EOF
```

Make a text file with a long secret:

<!-- @makeLongSecretFile @test -->
```
cat <<'EOF' >$DEMO_HOME/longsecret.txt
Lorem ipsum dolor sit amet,
consectetur adipiscing elit,
sed do eiusmod tempor incididunt
ut labore et dolore magna aliqua.
EOF
```

And make a kustomization file referring to the
above and additionally defining some literal KV
pairs:

<!-- @makeKustomization1 @test -->
```
cat <<'EOF' >$DEMO_HOME/kustomization.yaml
secretGenerator:
- name: mysecrets
  kvSources:
  - name: envfiles
    pluginType: builtin
    args:
    - foo.env
  - name: files
    pluginType: builtin 
    args:
    - longsecret.txt
  - name: literals
    pluginType: builtin
    args:
    - FRUIT=apple
    - VEGETABLE=carrot
EOF
```

> The above syntax is _alpha_ behavior at HEAD, for v2.1+.
>
> The default value of `pluginType` is `builtin`, so the
> `pluginType` fields could be omitted.
>
> The equivalent [v2.0.3] syntax (still supported) is
> ```
> secretGenerator:
> - name: mysecrets
>   env: foo.env
>   files:
>   - longsecret.txt
>   literals:
>   - FRUIT=apple
>   - VEGETABLE=carrot
> ```

Now generate the Secret:

<!-- @build1 @test -->
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

The name of the resource is a prefix, `mysecrets`
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

> New _alpha_ behavior at HEAD, for v2.1+

A general alternative is to enshrine secret
value generation in a [Go plugin].

The values can then come in via, say, an
authenticated and authorized RPC to a password
vault service.

Here's a trivial plugin that provides
hardcoded values:

<!-- @makePlugin @test -->
```
cat <<'EOF' >$DEMO_HOME/kvMaker.go
package main
var database = map[string]string{
  "TREE":      "oak",
  "ROCKET":    "Saturn V",
  "FRUIT":     "apple",
  "VEGETABLE": "carrot",
  "SIMPSON":   "homer",
}

type plugin struct{}
var KVSource plugin
func (p plugin) Get(
    root string, args []string) (map[string]string, error) {
  r := make(map[string]string)
  for _, k := range args {
    v, ok := database[k]
    if ok {
      r[k] = v
    }
  }
  return r, nil
}
EOF
```


The two crucial items needed to
load and query the plugin are
 1) the public symbol `KVSource`,
 1) its public `Get` method and signature.

Plugins that generate KV pairs for kustomize
must be installed at

> ```
> $XDG_CONFIG_HOME/kustomize/plugins/kvSource
> ```

`XDG_CONFIG_HOME` is an environment variable
used by many programs as the root of a
configuration directory.  If unspecified, the
default `$HOME/.config` is used.  The rest of
the required directory path establishes that
the files found there are kustomize plugins
for generating KV pairs.

Compile and install the plugin:

<!-- @compilePlugin @test -->
```
kvSources=$DEMO_HOME/kustomize/plugins/kvSources
mkdir -p $kvSources
go build -buildmode plugin \
  -o $kvSources/kvMaker.so \
  $DEMO_HOME/kvMaker.go
```

Create a new kustomization file
referencing this plugin:

<!-- @makeKustomization2 @test -->
```
cat <<'EOF' >$DEMO_HOME/kustomization.yaml
secretGenerator:
- name: mysecrets
  kvSources:
  - name: kvMaker
    pluginType: go
    args:
    - FRUIT
    - VEGETABLE
EOF
```

Finally, generate the secret, setting
`XDG_CONFIG_HOME` appropriately:

<!-- @build2 @test -->
```
result=$( \
  XDG_CONFIG_HOME=$DEMO_HOME \
  kustomize \
  --enable_alpha_goplugins_accept_panic_risk \
  build $DEMO_HOME )
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

Go plugins work well, but their usage may
fail (the program may crash) if there's
too much skew between _main program_ and
_plugin_ compilation conditions.  For
this reason, their use is protected by an
annoyingly long opt-in flag
(`--enable_alpha_goplugins_accept_panic_risk`)
intended to make the user aware of this risk.

It's safest to use Go plugins in the
context of a container image holding both
the main and the Go plugins it needs, all built
on the same machine, with the same transitive
libs and the same compiler version.
