#!/bin/bash

# Downloads the most recently released kustomize binary
# to your current working directory.
#
# Fails if the file already exists.

where=$PWD
if [ -f $where/kustomize ]; then
  echo "A file named kustomize already exists (remove it first)."
  exit 1
fi

tmpDir=`mktemp -d`
if [[ ! "$tmpDir" || ! -d "$tmpDir" ]]; then
  echo "Could not create temp dir."
  exit 1
fi

function cleanup {      
  rm -rf "$tmpDir"
}

trap cleanup EXIT

pushd $tmpDir >& /dev/null

opsys=windows
if [[ "$OSTYPE" == linux* ]]; then
  opsys=linux
elif [[ "$OSTYPE" == darwin* ]]; then
  opsys=darwin
fi

curl -s https://api.github.com/repos/kubernetes-sigs/kustomize/releases |\
  grep browser_download |\
  grep $opsys |\
  cut -d '"' -f 4 |\
  grep /kustomize/v |\
  sort | tail -n 1 |\
  xargs curl -s -O -L

tar xzf ./kustomize_v*_${opsys}_amd64.tar.gz

cp ./kustomize $where

popd >& /dev/null

./kustomize version

echo kustomize installed to current directory.
