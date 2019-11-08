#!/bin/bash

set -e
set -o xtrace

function clonePseudoRepo {
  git clone -b kubernetes-1.16.2 --depth 1 \
    https://github.com/kubernetes/$1.git
  rm -rf $1/.git
  find $1 -name go.mod | xargs rm
  find $1 -name go.sum | xargs rm
  find $1 -name OWNERS | xargs rm
}

function replacePseudoModuleName {
  find . -name *.go | xargs sed -i -e "s!k8s.io/$1!sigs.k8s.io/kustomize/pseudo/k8s/$1!g"
  find . -name *.proto | xargs sed -i -e "s!k8s.io/$1!sigs.k8s.io/kustomize/pseudo/k8s/$1!g"
}

function checkForForbiddenModules {
    if find . -name '*.go' | xargs grep "k8s.io/$1" ; then
      echo "forbidden dep k8s.io/$1 in *.go"
      exit 1
    fi
    if find . -name 'go.*' | xargs grep "k8s.io/$1" ; then
      echo "forbidden dep k8s.io/$1 in go.*"
      exit 1
    fi
}

# make sure we are running in the right spot
if [ ! -d "pseudo" ]; then
  echo "must run script from kustomize root dir"
  exit 1
fi
cd pseudo

# make the k8s deps dir
if [ -d "k8s" ]; then
  echo "must remove existing k8s dir"
  exit 1
fi
mkdir k8s

cd k8s
go mod init sigs.k8s.io/kustomize/pseudo/k8s

# setup the correct set of dependencies -- copied from the client-go repo
go mod edit \
  -require=github.com/google/go-cmp@v0.3.1 \
  -require=github.com/google/gofuzz@v1.0.0 \
  -require=github.com/google/uuid@v1.1.1 \
  -require=github.com/imdario/mergo@v0.3.5 \
  -require=github.com/json-iterator/go@v1.1.8 \
  -require=github.com/modern-go/reflect2@v1.0.1 \
	-require=golang.org/x/crypto@v0.0.0-20190308221718-c2843e01d9a2 \
  -require=golang.org/x/net@v0.0.0-20190620200207-3b0461eec859 \
  -require=golang.org/x/oauth2@v0.0.0-20190604053449-0f29369cfe45 \
  -require=golang.org/x/time@v0.0.0-20191024005414-555d28b269f0 \
  -require=gopkg.in/yaml.v2@v2.2.4 \
  -require=k8s.io/klog@v1.0.0 \
  -require=k8s.io/kube-openapi@v0.0.0-20191107075043-30be4d16710a \
  -require=k8s.io/utils@v0.0.0-20191030222137-2b95a09bc58d \
  -require=sigs.k8s.io/yaml@v1.1.0

# setup the correct set of dependencies -- copied from the client-go repo
go mod edit \
  -replace=golang.org/x/sys=golang.org/x/sys@v0.0.0-20190813064441-fde4db37ae7a \
  -replace=golang.org/x/tools=golang.org/x/tools@v0.0.0-20190821162956-65e3620a7ae7

# fetch the k8s packages
for item in api apimachinery client-go
do
  clonePseudoRepo $item
done

# fixup the package names
for item in api apimachinery client-go
do
  replacePseudoModuleName $item
done

# test the pseudo packages
go test ./...

# verify the package dependencies
for item in api apimachinery client-go
do
  checkForForbiddenModules $item
done
