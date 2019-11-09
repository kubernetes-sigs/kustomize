#!/bin/bash

set -o xtrace

for dir in api kustomize pseudo kyaml plugin cmd/kyaml
do
  for item in api apimachinery client-go
  do
    if find $dir -name 'go.*' | xargs grep "k8s.io/${item}" ; then
      echo "forbidden deps"
      exit 1
    fi
    if find $dir -name '*.go' | xargs grep "k8s.io/${item}" ; then
      echo "forbidden deps"
      exit 1
    fi
  done
done