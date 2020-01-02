# Qlik kustomize fork

From: https://github.com/kubernetes-sigs/kustomize

Contains Qlik plugins as part of the executable

## Building
```bash
git clone git@github.com:qlik-oss/kustomize.git kustomize_fork
cd kustomize_fork/kustomize
go install .
```

## Usage
```bash
$HOME/go/bin/kustomize build . --load_restrictor=none
```