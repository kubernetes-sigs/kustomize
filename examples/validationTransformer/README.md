# a transformer plugin performing validation

[base]: ../../docs/glossary.md#base
[kubeval]: https://github.com/instrumenta/kubeval
[plugin]: ../../docs/plugins

kustomize doesn't validate either its input or
output beyond the validation provided by the
marshalling/unmarshalling packages it depends on.

Another tool, [kubeval], goes beyond this to do
k8s aware validation. Here's a usage example:

```shell
$ kubeval my-invalid-rc.yaml
The document my-invalid-rc.yaml contains an invalid ReplicationController
--> spec.replicas: Invalid type. Expected: integer, given: string
```

One can write a Kustomize transformer [plugin] to
run [kubeval] against the resources that have been
loaded by Kustomize.


Make a place to work:

<!-- @makeWorkplace @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)
mkdir -p $DEMO_HOME/valid
mkdir -p $DEMO_HOME/invalid
PLUGINDIR=$DEMO_HOME/kustomize/plugin/someteam.example.com/v1/validator
mkdir -p $PLUGINDIR
```

## write a transformer plugin

Download the [kubeval] binary depending on the operating system
and add it to $PATH.

<!-- @downloadKubeval @testAgainstLatestRelease -->
```
OS=`uname | sed -e 's/Linux/linux/' -e 's/Darwin/darwin/'`
wget https://github.com/instrumenta/kubeval/releases/download/0.9.2/kubeval-${OS}-amd64.tar.gz
tar xf kubeval-${OS}-amd64.tar.gz
export PATH=$PATH:`pwd`
```

Kustomize has the following assumption of a transformer plugin:
- The resources are passed to the transformer plugin from stdin.
- The configuration file for the transformer plugin is passed in
  as the first argument.
- The working directory of the plugin is the kustomization
  directory where it is used as a transformer.
- The transformed resources are written to stdout by the plugin.
- If the return code of the transformer plugin is non zero,
  Kustomize regards there is an error during the transformation.

A transformer plugin for the validation can be written as a
bash script, which execute the [kubeval] binary and return proper
output and exit code.

<!-- @writePlugin @testAgainstLatestRelease -->
```
cat <<'EOF' > $PLUGINDIR/Validator
#!/bin/bash

if ! [ -x "$(command -v kubeval)" ]; then
  echo "Error: kubeval is not installed."
  exit 1
fi

temp_file=$(mktemp)
output_file=$(mktemp)
cat - > $temp_file

kubeval $temp_file > $output_file

if [ $? -eq 0 ]; then
    cat $temp_file
    rm $temp_file $output_file
    exit 0
fi

cat $output_file
rm $temp_file $output_file
exit 1

EOF
chmod +x $PLUGINDIR/Validator
```

## use the transformer plugin

Define a kustomization containing a valid ConfigMap
and the transformer plugin.

<!-- @writeKustomization @testAgainstLatestRelease -->
```
cat <<'EOF' >$DEMO_HOME/valid/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  foo: bar
EOF

cat <<'EOF' >$DEMO_HOME/valid/validation.yaml
apiVersion: someteam.example.com/v1
kind: Validator
metadata:
  name: notImportantHere
EOF

cat <<'EOF' >$DEMO_HOME/valid/kustomization.yaml
resources:
- configmap.yaml

transformers:
- validation.yaml
EOF
```

Define a kustomization containing an invalid ConfigMap
and the transformer plugin.

<!-- @writeKustomization @testAgainstLatestRelease -->
```
cat <<'EOF' >$DEMO_HOME/invalid/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
- foo: bar
EOF

cat <<'EOF' >$DEMO_HOME/invalid/validation.yaml
apiVersion: someteam.example.com/v1
kind: Validator
metadata:
  name: notImportantHere
EOF

cat <<'EOF' >$DEMO_HOME/invalid/kustomization.yaml
resources:
- configmap.yaml

transformers:
- validation.yaml
EOF
```

The directory structure is as the following:

```
/tmp/tmp.fAYMfLZJs4
├── invalid
│   ├── configmap.yaml
│   ├── kustomization.yaml
│   └── validation.yaml
├── kustomize
│   └── plugin
│       └── someteam.example.com
│           └── v1
│               ├── kubeval
│               └── Validator
└── valid
    ├── configmap.yaml
    ├── kustomization.yaml
    └── validation.yaml
```

Define a helper function to run kustomize with the
correct environment and flags for plugins:

<!-- @defineKustomizeBd @testAgainstLatestRelease -->
```
function kustomizeBd {
  XDG_CONFIG_HOME=$DEMO_HOME \
  kustomize build \
    --enable_alpha_plugins \
    $DEMO_HOME/$1
}
```

Build the valid variant

<!-- @buildValid @testAgainstLatestRelease -->
```
kustomizeBd valid
```
The output contains a ConfigMap as

```yaml
apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  name: cm
```

Build the invalid variant

```
kustomizeBd invalid
```

The output is an error as
```shell
data: Invalid type. Expected: object, given: array
```

## cleanup

<!-- @cleanup @testAgainstLatestRelease -->
```shell
rm -rf $DEMO_HOME
```
