#!/bin/bash

set -o xtrace

for item in api apimachinery client-go
do
  if find pseudo -name 'go.*' | xargs grep "k8s.io/${item}" ; then
    echo "forbidden deps"
    exit 1
  fi
  if find pseudo -name '*.go' | xargs grep "k8s.io/${item}" ; then
    echo "forbidden deps"
    exit 1
  fi
done

