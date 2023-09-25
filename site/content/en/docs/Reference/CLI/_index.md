---
title: "CLI"
linkTitle: "CLI"
weight: 3
date: 2023-07-28
description: >
  Reference for the Command Line Interface.
---

This overview covers `kustomize` syntax, describes the command operations, and provides common examples.

## Syntax
Use the following syntax to run `kustomize` commands from your terminal window:

```bash
kustomize [command]
```

The `command` flag specifies the operation that you want to perform, for example `create`, `build`, `cfg`.

If you need help, run `kustomize help` from the terminal window.

## Operations
The following table includes short descriptions and the general syntax for all the `kustomize` operations.

Operation | Syntax | Description
--- | --- | ---
build | `kustomize build DIR [flags]` | Build a kustomization target from a directory or URL.
cfg | `kustomize cfg [command]` | Commands for reading and writing configuration.
completion | `kustomize completion` [bash\|zsh\|fish\|powershell] | Generate shell completion script.
create | `kustomize create [flags]` | Create a new kustomization in the current directory.
edit | `kustomize edit [command]` |  Edits a kustomization file.
fn | `kustomize fn [command]` | Commands for running functions against configuration.
localize | `kustomize localize [target [destination]] [flags]` | [Alpha] Creates localized copy of target kustomization root at destination.
version | `kustomize version [flags]` | Prints the kustomize version.

## Examples: Common Operations
Use the following set of examples to help you familiarize yourself with running the commonly used `kustomize` operations:

`kustomize build` - Build a kustomization target from a directory or URL.

```bash
# Build the current working directory
kustomize build

# Build some shared configuration directory
kustomize build /home/config/production

# Build from github
kustomize build https://github.com/kubernetes-sigs/kustomize.git/examples/helloWorld?ref=v1.0.6
```

`kustomize create` - Create a new kustomization in the current directory.
```bash
# Create an empty kustomization.yaml file
kustomize create

# Create a new overlay from the base '../base".
kustomize create --resources ../base

# Create a new kustomization detecting resources in the current directory.
kustomize create --autodetect

# Create a new kustomization with multiple resources and fields set.
kustomize create --resources deployment.yaml,service.yaml,../base --namespace staging --nameprefix acme-
```

`kustomize edit` - Edits a kustomization file.
```bash
# Adds a configmap to the kustomization file
kustomize edit add configmap NAME --from-literal=k=v

# Sets the nameprefix field
kustomize edit set nameprefix <prefix-value>

# Sets the namesuffix field
kustomize edit set namesuffix <suffix-value>
```
