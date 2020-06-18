# Examples for Validator Plugin

Previously, Kustomize suggested to used a transformer plugin to [perform validation](https://github.com/kubernetes-sigs/kustomize/tree/master/examples/validationTransformer). Now we introduce a new type of plugin: validator. As the name says, validator is used to validate the result YAML output. It works in the same way with *transformers* but cannot *modify* the input YAML content. Let's take a look at how it works.

## Make a Place to Work

<!-- @makeWorkplace @validatorPlugin -->
```
DEMO_HOME=$(mktemp -d)
mkdir -p $DEMO_HOME/valid
PLUGINDIR=$DEMO_HOME/kustomize/plugin/someteam.example.com/v1/validator
mkdir -p $PLUGINDIR
```

## Write a Validator Plugin

Kustomize has the following assumption of a validator plugin:
- The resources are passed to the validator plugin from stdin.
- The configuration file for the validator plugin is passed in
  as the first argument.
- The working directory of the plugin is the kustomization
  directory where it is used as a validator.
- The validated resources are written to stdout by the plugin. Or the validator can print nothing to the stdout if there is no need to change the input.
- Validator can **only** add a label named `validated-by` (case-sensitive) to the **top-level** resources. If there is any other modification in the validator, Kustomize will throw an error.
- If the return code of the transformer plugin is non zero,
  Kustomize regards there is an error during the validation.

You can use either exec plugin or Go plugin as a validator. Here we use a bash script as an exec plugin.

<!-- @writePlugin @validatorPlugin -->
```bash
cat <<'EOF' > $PLUGINDIR/Validator
#!/bin/bash

# Do whatever you want here. In this example we
# just print out the input

cat

EOF
chmod +x $PLUGINDIR/Validator
```

## Use the Validator Plugin

Define a kustomization containing a valid ConfigMap
and the transformer plugin.

<!-- @writeKustomization @validatorPlugin -->
```bash
cat <<'EOF' >$DEMO_HOME/valid/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  foo: bar
EOF

cat <<'EOF' >$DEMO_HOME/valid/validator.yaml
apiVersion: someteam.example.com/v1
kind: Validator
metadata:
  name: notImportantHere
EOF

cat <<'EOF' >$DEMO_HOME/valid/kustomization.yaml
resources:
- configmap.yaml

validators:
- validator.yaml
EOF
```

The directory structure is as the following:

```
/tmp/tmp.69tTCuXuYc
├── kustomize
│   └── plugin
│       └── someteam.example.com
│           └── v1
│               └── validator
│                   └── Validator
└── valid
    ├── configmap.yaml
    ├── kustomization.yaml
    └── validator.yaml
```

Define a helper function to run kustomize with the
correct environment and flags for plugins:

<!-- @defineKustomizeBd @validatorPlugin -->
```bash
function kustomizeBd {
  XDG_CONFIG_HOME=$DEMO_HOME \
  kustomize build \
    --enable_alpha_plugins \
    $DEMO_HOME/$1
}
```

Build the valid variant

<!-- @buildValid @validatorPlugin -->
```bash
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

### Validator Failure

Now lets try a failed validator

```bash
cat <<'EOF' > $PLUGINDIR/Validator
#!/bin/bash

# Non-zero indicates a failed validation
>&2 echo "Validation failed"
exit 1

EOF
chmod +x $PLUGINDIR/Validator
```

Build the valid variant

```bash
kustomizeBd valid
```
The output contains the error information that is printed to stderr
by validator.

```
Validation failed
Error: failure in plugin configured via /tmp/kust-plugin-config-369137659; exit status 1: exit status 1
```

### Input Modification

Typically a validator shouldn't modify the content to be validated. If it does, Kustomize will complain about it.

```bash
cat <<'EOF' > $PLUGINDIR/Validator
#!/bin/bash

# Modify the input content

sed 's/bar/baz/g'

EOF
chmod +x $PLUGINDIR/Validator
```

Then build

```
kustomizeBd valid
```

The error output will indicate you where is modified by the validator

```
Error: validator shouldn't modify the resource map: kunstruct not equal:
 -- {"apiVersion":"v1","data":{"foo":"bar"},"kind":"ConfigMap","metadata":{"name":"cm"}}{nsfx:false,beh:unspecified},
 -- {"apiVersion":"v1","data":{"foo":"baz"},"kind":"ConfigMap","metadata":{"name":"cm"}}{nsfx:false,beh:unspecified}

--
&resource.Resource{Kunstructured:(*kunstruct.UnstructAdapter)(0xc000118408), originalName:"cm", originalNs:"", options:(*types.GenArgs)(0xc00059e5e8), refBy:[]resid.ResId(nil), refVarNames:[]string(nil), namePrefixes:[]string{""}, nameSuffixes:[]string{""}}
------
&resource.Resource{Kunstructured:(*kunstruct.UnstructAdapter)(0xc000118510), originalName:"cm", originalNs:"", options:(*types.GenArgs)(0xc00059e5e8), refBy:[]resid.ResId(nil), refVarNames:[]string(nil), namePrefixes:[]string{""}, nameSuffixes:[]string{""}}
```

There is an exception that the validator can add a `validated-by` label to the **top** level resources.

<!-- @validatedByLabel @validatorPlugin -->
```bash
cat <<'EOF' > $PLUGINDIR/Validator
#!/usr/bin/bash

sed 's/^  name: cm$/  name: cm\n  labels:\n    validated-by: whatever/'

EOF
chmod +x $PLUGINDIR/Validator
```

Then build

<!-- @validatedByLabelBuild @validatorPlugin -->
```
kustomizeBd valid
```

The output will be

```yaml
apiVersion: v1
data:
  foo: bar
kind: ConfigMap
metadata:
  labels:
    validated-by: whatever
  name: cm
```

## cleanup

<!-- @cleanup @validatorPlugin -->
```
rm -rf $DEMO_HOME
```