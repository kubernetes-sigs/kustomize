#!/bin/bash

# If no argument is given -> Downloads the most recently released
# kustomize binary to your current working directory.
# (e.g. 'install_kustomize.sh')
#
# If an argument is given -> Downloads the specified version of the
# kustomize binary to your current working directory.
# (e.g. 'install_kustomize.sh 3.8.2')
#
# Fails if the file already exists.

if [ -z "$1" ]
  then
    echo "No version specified. Downloading the most recently released kustomize binary."
    version=""
  else
    echo "Downloading the kustomize binary version $1."
    version=$1
fi

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
  grep /kustomize/v$version |\
  sort | tail -n 1 |\
  xargs curl -s -O -L

if [ -e ./kustomize_v*_${opsys}_amd64.tar.gz ]
then
    tar xzf ./kustomize_v*_${opsys}_amd64.tar.gz
else
    echo "Error: kustomize binary with the version $version does not exist!"
    exit 1
fi

cp ./kustomize $where

popd >& /dev/null

./kustomize version

echo kustomize installed to current directory.
