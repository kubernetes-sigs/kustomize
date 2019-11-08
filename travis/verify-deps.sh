#!/bin/bash

if find api forked -name go.sum | xargs grep -E "k8s\.io/(api|client-go)" ; then
  echo "deps not allowed"
  find api forked -name go.sum | xargs grep -E "k8s\.io/(api|client-go)"
  exit 1
fi
