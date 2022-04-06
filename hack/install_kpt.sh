#!/bin/bash
# Copyright 2022 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0


# If no arguments are given -> Downloads the most recently released
# kpt binary to your current working directory.
# (e.g. 'install_kpt.sh')
#
# If one argument is given -> Downloads the specified version of the
# kpt binary to your current working directory.
# (e.g. 'install_kpt.sh 0.34.0')
#
# If two arguments are given -> Downloads the specified version of the
# kpt binary to the specified directory.
# (e.g. 'install_kpt.sh 0.34.0 $(go env GOPATH)/bin')
#
# Fails if the file already exists.

if [ -z "$1" ]; then
    version=""
  else
    version=$1
fi

if [ -z "$2" ]; then
    where=$PWD
  else
    where=$2
fi

if [ -f $where/kpt ]; then
  echo "A file named kpt already exists (remove it first)."
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

curl -s https://api.github.com/repos/GoogleContainerTools/kpt/releases |\
  grep browser_download |\
  grep $opsys |\
  cut -d '"' -f 4 |\
  grep /kpt/releases/download/v$version |\
  sort | tail -n 1 |\
  xargs curl -s -O -L

if [ -e ./kpt_${opsys}_amd64-*.tar.gz ]; then
    tar xzf ./kpt_${opsys}_amd64-*.tar.gz
else
    echo "Error: kpt binary with the version $version does not exist!"
    exit 1
fi

cp ./kpt $where

popd >& /dev/null

$where/kpt version

echo kpt installed to specified directory.
