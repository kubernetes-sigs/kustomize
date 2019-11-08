#!/bin/bash

set -e
set -o xtrace

# replace the module name with the new module name
function replaceModuleName {
  find . -name *.go | xargs sed -i -e "s!k8s.io/$1!sigs.k8s.io/kustomize/pseudo/k8s/$1!g"
  find . -name *.proto | xargs sed -i -e "s!k8s.io/$1!sigs.k8s.io/kustomize/pseudo/k8s/$1!g"
  find . -name *.md | xargs sed -i -e "s!k8s.io/$1!sigs.k8s.io/kustomize/pseudo/k8s/$1!g"
}

# update the go.mod file, dropping the old module
function updateGoModFile {
  go mod edit -droprequire=k8s.io/$1 || echo ""
  rm go.sum || echo ""
}

# test the module
function testGoMod {
  go test ./...
  gofmt -s -w .
  go mod tidy
}

# update all go.mod files
function updateAllGoModFiles {
  (cd api; updateGoModFile $1 )
  (cd kustomize; updateGoModFile $1 )
  (cd hack/crawl; updateGoModFile $1 )
  (cd pluginator; updateGoModFile $1 )

  for goMod in $(find ./plugin/builtin -name 'go.mod'); do
    dir=$(dirname "${goMod}")
    (cd $dir; updateGoModFile $1 )
  done

  for goMod in $(find ./plugin/someteam.example.com/v1 -name 'go.mod'); do
    dir=$(dirname "${goMod}")
    (cd $dir; updateGoModFile $1 )
  done
}

# test all go modules
function testAllModules {
  (cd api; testGoMod)
  (cd kustomize; testGoMod)
  (cd hack/crawl; testGoMod)
  (cd pluginator; testGoMod )

  for goMod in $(find ./plugin/builtin -name 'go.mod'); do
    dir=$(dirname "${goMod}")
    (cd $dir; testGoMod )
  done

#  Uncomment this when tests are added for this module
#  for goMod in $(find ./plugin/someteam.example.com/v1 -name 'go.mod'); do
#    dir=$(dirname "${goMod}")
#    (cd $dir; testGoMod )
#  done
}

# update the names of the modules
for item in api apimachinery client-go
do
  replaceModuleName $item
done

# update the go.mod files
for item in api apimachinery client-go
do
  updateAllGoModFiles $item
done

# test all of the modules still work
testAllModules

./travis/verify-deps.sh