# Go Plugin Guided Example for Linux

[SopsEncodedSecrets repository]: https://github.com/monopole/sopsencodedsecrets
[Go plugin]: https://golang.org/pkg/plugin
[Go plugin caveats]: goPluginCaveats.md

This is a (no reading allowed!) 60 second copy/paste guided
example.

Full plugin docs [here](README.md).
Be sure to read the [Go plugin caveats].

This demo uses a Go plugin, `SopsEncodedSecrets`,
that lives in the [sopsencodedsecrets repository].
This is an inprocess [Go plugin], not an
sub-process exec plugin that happens to be written
in Go (which is another option for Go authors).

This is a guide to try it without damaging your
current setup.

#### requirements

* linux, git, curl, Go 1.12

For encryption

* gpg

Or

* Google cloud (gcloud) install
* a Google account with KMS permission

## Make a place to work

```shell
# Keeping these separate to avoid cluttering the DEMO dir.
DEMO=$(mktemp -d)
tmpGoPath=$(mktemp -d)
```

## Install kustomize

Need v3.0.0 for what follows, and you must _compile_
it (not download the binary from the release page):

```shell
GOPATH=$tmpGoPath go install sigs.k8s.io/kustomize/v3/cmd/kustomize
```

## Make a home for plugins

A kustomize plugin is fully determined by
its configuration file and source code.

[required fields]: https://kubernetes.io/docs/concepts/overview/working-with-objects/kubernetes-objects/#required-fields

Kustomize plugin configuration files are formatted
as kubernetes resource objects, meaning
`apiVersion`, `kind` and `metadata` are [required
fields] in these config files.

The kustomize program reads the config file
(because the config file name appears in the
`generators` or `transformers` field in the
kustomization file), then locates the Go plugin's
object code at the following location:

> ```shell
> $XGD_CONFIG_HOME/kustomize/plugin/$apiVersion/$lKind/$kind.so
> ```

where `lKind` holds the lowercased kind.  The
plugin is then loaded and fed its config, and the
plugin's output becomes part of the overall
`kustomize build` process.

The same plugin might be used multiple times in
one kustomize build, but with different config
files.  Also, kustomize might customize config
data before sending it to the plugin, for whatever
reason.  For these reasons, kustomize owns the
mapping between plugins and config data; it's not
left to plugins to find their own config.

This demo will house the plugin it uses at the
ephemeral directory

```shell
PLUGIN_ROOT=$DEMO/kustomize/plugin
```

and ephemerally set `XGD_CONFIG_HOME` on a command
line below.

### What apiVersion and kind?

At this stage in the development of kustomize
plugins, plugin code doesn't know or care what
`apiVersion` or `kind` appears in the config file
sent to it.

The plugin could check these fields, but it's the
remaining fields that provide actual configuration
data, and at this point the successful parsing of
these other fields are the only thing that matters
to a plugin.

This demo uses a plugin called _SopsEncodedSecrets_,
and it lives in the [SopsEncodedSecrets repository].

Somewhat arbitrarily, we'll chose to install
this plugin with

```shell
apiVersion=mygenerators
kind=SopsEncodedSecrets
```

### Define the plugin's home dir

By convention, the ultimate home of the plugin
code and supplemental data, tests, documentation,
etc. is the lowercase form of its kind.

```shell
lKind=$(echo $kind | awk '{print tolower($0)}')
```

### Download the SopsEncodedSecrets plugin

In this case, the repo name matches the lowercase
kind already, so we just clone the repo and get
the proper directory name automatically:

```shell
mkdir -p $PLUGIN_ROOT/${apiVersion}
cd $PLUGIN_ROOT/${apiVersion}
git clone git@github.com:monopole/sopsencodedsecrets.git
```

Remember this directory:

```shell
MY_PLUGIN_DIR=$PLUGIN_ROOT/${apiVersion}/${lKind}
```

### Try the plugin's own test

Plugins may come with their own tests.
This one does, and it hopefully passes:

```shell
cd $MY_PLUGIN_DIR
go test SopsEncodedSecrets_test.go
```

Build the object code for use by kustomize:

```shell
cd $MY_PLUGIN_DIR
GOPATH=$tmpGoPath go build -buildmode plugin -o ${kind}.so ${kind}.go
```

This step may succeed, but kustomize might
ultimately fail to load the plugin because of
dependency [skew].

[skew]: https://github.com/kubernetes-sigs/kustomize/blob/master/docs/plugins/README.md#caveats
[used in this demo]: #install-kustomize

On load failure

 * be sure to build the plugin with the same
   version of Go (_go1.12_) on the same `$GOOS`
   (_linux_) and `$GOARCH` (_amd64_) used to build
   the kustomize being [used in this demo].

 * change the plugin's dependencies in its `go.mod`
   to match the versions used by kustomize (check
   kustomize's `go.mod` used in its tagged commit).

Lacking tools and metadata to allow this to be
automated, there won't be a Go plugin ecosystem.

Kustomize has adopted a Go plugin architecture as
to ease accept new generators and transformers
(just write a plugin), and to be sure that native
operations (also constructed and tested as
plugins) are compartmentalized, orderable and
reusable instead of bizarrely woven throughout the
code as a individual special cases.

## Create a kustomization

Make a kustomization directory to
hold all your config:

```shell
MYAPP=$DEMO/myapp
mkdir -p $MYAPP
```

Make a config file for the SopsEncodedSecrets plugin.

Its `apiVersion` and `kind` allow the plugin to be
found:

```shell
cat <<EOF >$MYAPP/secGenerator.yaml
apiVersion: ${apiVersion}
kind: ${kind}
metadata:
  name: mySecretGenerator
name: forbiddenValues
namespace: production
file: myEncryptedData.yaml
keys:
- ROCKET
- CAR
EOF
```

This plugin expects to find more data in
`myEncryptedData.yaml`; we'll get to that shortly.

Make a kustomization file referencing the plugin
config:

```shell
cat <<EOF >$MYAPP/kustomization.yaml
commonLabels:
  app: hello
generators:
- secGenerator.yaml
EOF
```

Now generate the real encrypted data.

### Assure you have an encryption tool installed

We're going to use [sops](https://github.com/mozilla/sops) to encode a file. Choose either GPG or Google Cloud KMS as the secret provider to continue.

#### GPG

Try this:

```shell
gpg --list-keys
```

If it returns a list, presumably you've already created keys. If not, try import test keys from sops for dev.

```shell
curl https://raw.githubusercontent.com/mozilla/sops/master/pgp/sops_functional_tests_key.asc | gpg --import
SOPS_PGP_FP="1022470DE3F0BC54BC6AB62DE05550BC07FB1A0A"
```

#### Google Cloude KMS

Try this:

```shell
gcloud kms keys list --location global --keyring sops
```

If it succeeds, presumably you've already created keys and placed them in a keyring called sops. If not, do this:

```shell
gcloud kms keyrings create sops --location global
gcloud kms keys create sops-key --location global \
    --keyring sops --purpose encryption
```

Extract your keyLocation for use below:

```shell
keyLocation=$(\
    gcloud kms keys list --location global --keyring sops |\
    grep GOOGLE | cut -d " " -f1)
echo $keyLocation
```

### Install `sops`

```shell
GOPATH=$tmpGoPath go install go.mozilla.org/sops/cmd/sops
```

### Create data encrypted with your private key

Create raw data to encrypt:

```shell
cat <<EOF >$MYAPP/myClearData.yaml
VEGETABLE: carrot
ROCKET: saturn-v
FRUIT: apple
CAR: dymaxion
EOF
```

Encrypt the data into file the plugin wants to read:

With PGP

```shell
$tmpGoPath/bin/sops --encrypt \
  --pgp $SOPS_PGP_FP \
  $MYAPP/myClearData.yaml >$MYAPP/myEncryptedData.yaml
```

Or GCP KMS

```shell
$tmpGoPath/bin/sops --encrypt \
  --gcp-kms $keyLocation \
  $MYAPP/myClearData.yaml >$MYAPP/myEncryptedData.yaml
```

Review the files

```shell
tree $DEMO
```

This should look something like:

> ```shell
> /tmp/tmp.0kIE9VclPt
> ├── kustomize
> │   └── plugin
> │       └── mygenerators
> │           └── sopsencodedsecrets
> │               ├── go.mod
> │               ├── go.sum
> │               ├── LICENSE
> │               ├── README.md
> │               ├── SopsEncodedSecrets.go
> │               ├── SopsEncodedSecrets.so
> │               └── SopsEncodedSecrets_test.go
> └── myapp
>     ├── kustomization.yaml
>     ├── myClearData.yaml
>     ├── myEncryptedData.yaml
>     └── secGenerator.yaml
> ```

## Build your app, using the plugin:

```shell
XDG_CONFIG_HOME=$DEMO $tmpGoPath/bin/kustomize build --enable_alpha_plugins $MYAPP
```

This should emit a kubernetes secret, with
encrypted data for the names `ROCKET` and `CAR`.

Above, if you had set

> ```shell
> PLUGIN_ROOT=$HOME/.config/kustomize/plugin
> ```

there would be no need to use `XDG_CONFIG_HOME` in the
_kustomize_ command above.
