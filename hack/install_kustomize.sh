#!/usr/bin/env bash

# If no argument is given -> Downloads the most recently released
# kustomize binary to your current working directory.
# (e.g. 'install_kustomize.sh')
#
# If one argument is given -> 
# If that argument is in the format of #.#.#, downloads the specified
# version of the kustomize binary to your current working directory.
# If that argument is something else, downloads the most recently released
# kustomize binary to the specified directory.
# (e.g. 'install_kustomize.sh 3.8.2' or 'install_kustomize.sh $(go env GOPATH)/bin')
#
# If two arguments are given -> Downloads the specified version of the
# kustomize binary to the specified directory.
# (e.g. 'install_kustomize.sh 3.8.2 $(go env GOPATH)/bin')
#
# Fails if the file already exists.

set -e

curl_timeout=600

where=$PWD

version=""
release_url=https://api.github.com/repos/kubernetes-sigs/kustomize/releases
if [ -n "$1" ]; then
  if [[ "$1" =~ ^[0-9]+(\.[0-9]+){2}$ ]]; then
    version=v$1
    release_url=${release_url}/tags/kustomize%2F$version
  elif [ -n "$2" ]; then
    echo "The first argument should be the requested version."
    exit 1
  else
    where="$1"
  fi
fi

if [ -n "$2" ]; then
  where="$2"
fi

if ! test -d "$where"; then
  echo "$where does not exist. Create it first."
  exit 1
fi

where="$(readlink -f $where)/"

if [ -f "${where}kustomize" ]; then
  echo "A file named ${where}kustomize already exists (remove it first)."
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

trap cleanup EXIT ERR

pushd "$tmpDir" >& /dev/null

opsys=windows
arch=amd64
if [[ "$OSTYPE" == linux* ]]; then
  opsys=linux
elif [[ "$OSTYPE" == darwin* ]]; then
  opsys=darwin
fi

curl -m $curl_timeout -s $release_url |\
  grep browser_download.*${opsys}_${arch} |\
  cut -d '"' -f 4 |\
  sort -V | tail -n 1 |\
  xargs curl -m $curl_timeout -sLO

if [ -e ./kustomize_v*_${opsys}_amd64.tar.gz ]; then
    tar xzf ./kustomize_v*_${opsys}_amd64.tar.gz
else
    echo "Error: kustomize binary with the version ${version#v} does not exist!"
    exit 1
fi

cp ./kustomize "$where"

popd >& /dev/null

${where}kustomize version

echo "kustomize installed to ${where}kustomize"
