#!/bin/bash
#
# Converts Go-based kustomize plugins in
#   sigs.k8s.io/kustomize/plugin/builtin
# (all in package 'main') to generator and
# transformer factory functions in
#   sigs.k8s.io/kustomize/plugin/builtingen
# (all in package 'builtingen').
#
# Cannot put all these in the same dir, since
# plugins must be in the 'main' package,
# the factory functions cannot be in 'main',
# Go disallows multiple packages in one dir.

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

/bin/rm -rf plugin/builtingen
mkdir plugin/builtingen
GOPATH=$myGoPath go generate --tags plugin \
    sigs.k8s.io/kustomize/plugin/builtin
ls -C1 plugin/builtingen

popd >& /dev/null

echo All done.
