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
honored by many programs as the root of a
[configuration directory].  If the variable is
undefined, the convention is to fall back to
`$HOME/.config`.

The rest of the required directory path
establishes that the files found there are
kustomize plugins for generating KV pairs.

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
`XDG_CONFIG_HOME` so that the plugin
can be found under `$DEMO_HOME`:

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

### Go Plugin Caveats

Kustomize supports Go plugins to allow someone to
extend kustomize in type-safe fashion against a
documented Go interface type, without having to
get their code merged into the kustomize
repository, and without having to maintain a
permanent fork.

Go plugins work well, but fall short of what many
people think of when they hear the word _plugin_.

Go plugin compilation creates an [ELF] formatted
`.so` file, which by definition has no information
about the _provenance_ of the file.  One cannot
know which version of Go was used, which packages
were imported (and their version), what value of
`GOOS` and `GOARCH` were used, etc. If the skew
between the compilation conditions of the main
program ELF and the plugin ELF are too great, the
program will crash.  Also, there's no certificate
to check in a `.so` file, so no way to know who
wrote it or what it does.  A bare `.so` file, not
packaged with provenance information, is not a
suitable distrubution format.  It's not what
people expect from decades of adding features
IDEs, browsers, CAD tools, graphics tools, etc.
via things called _plugins_.

There's no reason why someone couldn't build a
`.so` packaging mechanism into `go` to emit an ELF
packaged with provenance allowing ELF
compatibility checks, but this isn't supported in
kustomize (or Go) at the time of writing.

To avoid provenance issues simply compile your Go
plugins and the main program at the same time.
Bundle them into a container image for use by
downstream users and/or your continuous delivery
bot.  This is the intended usage idiom for Go
plugins.

A `kustomize build` attempt with Go plugins that omits
the flag

> `--enable_alpha_goplugins_accept_panic_risk`

will fail with an error message about skew risks.
Flag use is an opt-in acknowledging the absence of
`.so` provenance, an absence that doesn't matter
to someone building the code from source.


### Leveraging Go plugins to run non-Go code


#### external services

For particular (user-created) transformations or
generations, kustomize could prepare a request,
send it to some service, and process a response.
How to do this is a [solved problem][grpc].  The
communication is struct-to-struct type safe - no
need to write parsing code.

If the service is written in Go, and one can
vendor its code, it's simplest to write a small Go
plugin that calls it like a library rather than
running the service as an independent process.

If the service is not written in Go, or if the
source code is unavailable, one can use a small Go
plugin to make the RPC.


#### subprocesses (also known as `exec` plugins)

In this approach one arranges for executable files
to be identified by name or location, and runs
them as a kustomize subprocess, sending a
'request' to the subprocess its `stdin`, and
obtaining a 'response' via its `stdout`.

An immediate way to use an arbitrary executable
with arbitrary i/o requirements is through a Go
plugin that runs the executable via
[`exec.Command`].  Each special purpose
tranformation or generation - needed by `kustomize
build` - will require it's own `stdin`/`stdout`
processing to convert from/to the Go types that
kustomize uses.

The Go plugin provides this translation layer, and
handles process exit codes.
