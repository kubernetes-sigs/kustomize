#!/bin/bash
set -e

# Builds or removes Go plugin object code.
#
# Specify plugin root as first arg, e.g.
#
#   ./hack/buildExternalGoPlugins.sh $KUSTOMIZE_PLUGIN_HOME
#   ./hack/buildExternalGoPlugins.sh $XDG_CONFIG_HOME/kustomize/plugin
#   ./hack/buildExternalGoPlugins.sh ${HOME}/.config/kustomize/plugin
#   ./hack/buildExternalGoPlugins.sh ./plugin
#
# add 2nd arg "clean" to remove instead of build.

root=$1
if [ ! -d $root ]; then
  echo "Don't see directory $root."
  exit 1
fi

fn=buildPlugin
if [ "$2" == "clean" ]; then
  fn=removePlugin
fi

function buildPlugin {
  echo "Building $1/$2.so"
  # Change dir so local go.mod applies
  pushd $1 >& /dev/null
  go build -buildmode plugin -o $2.so $2.go
  popd >& /dev/null
}

function removePlugin {
  local f="$1/$2.so"
  if [ -f "$f" ]; then
    echo "Removing $f"
    rm "$f"
  fi
}

goPlugins=$(
  find $root -name "*.go" |
  grep -v builtin/ |
  xargs grep -l "var KustomizePlugin")

for p in $goPlugins; do
  d=$(dirname "$p")
  n=$(basename "$p" | cut -f 1 -d '.')
  $fn $d $n
done
