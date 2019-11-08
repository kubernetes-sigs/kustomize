#!/usr/bin/env bash

# make local forks of k8s.io/api, k8s.io/apimachinery, and k8s.io/client-go

function cloneIt {
  git clone -b kubernetes-1.16.2 --depth 1 \
    https://github.com/kubernetes/$1.git
}

function makeClones {
  pushd ./api/internal

  rm -rf forked
  mkdir forked
  cd forked

  me=`basename "$0"`

  cat <<EOF >README.md

Code below this point was created by $me

Do not edit.

EOF

  cloneIt api
  cloneIt apimachinery
  cloneIt client-go

  popd
}

function fixGoMod {
  go mod edit \
    -require=k8s.io/api@v0.0.0 \
    -require=k8s.io/apimachinery@v0.0.0 \
    -require=k8s.io/client-go@v0.0.0
  
  go mod edit \
    -replace=k8s.io/api@v0.0.0=$1/internal/forked/api \
    -replace=k8s.io/apimachinery@v0.0.0=$1/internal/forked/apimachinery \
    -replace=k8s.io/client-go@v0.0.0=$1/internal/forked/client-go

  go mod tidy
}

function fixAllGoMods {
  (cd api; fixGoMod . )
  (cd kustomize; fixGoMod ../api )

  for goMod in $(find ./plugin/builtin -name 'go.mod'); do
    dir=$(dirname "${goMod}")
    (cd $dir; fixGoMod ../../../api )
  done

  for goMod in $(find ./plugin/someteam.example.com/v1 -name 'go.mod'); do
    dir=$(dirname "${goMod}")
    (cd $dir; fixGoMod ../../../../api )
  done
}
 
makeClones
fixAllGoMods


