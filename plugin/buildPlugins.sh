#!/bin/bash

set -e

myGoPath=$1
if [ -z ${1+x} ]; then
  myGoPath=$GOPATH
fi

if [ -z "$myGoPath" ]; then
  echo "Must specify a GOPATH"
  exit 1
fi

dir=$myGoPath/src/sigs.k8s.io/kustomize

if [ ! -d "$dir" ]; then
  echo "$dir is not a directory."
  exit 1
fi

echo Generating linkable plugins...

pushd $dir >& /dev/null

for i in `cat plugin/pluginList | grep -v '#'`
do
  echo $i
  GOPATH=$myGoPath GO111MODULE=on go build -buildmode plugin -o ${HOME}/.config/kustomize/${i}.so ./$i.go
done

popd >& /dev/null

echo All done.

