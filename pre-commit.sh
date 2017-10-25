#!/bin/bash

if [ -n "$failIt" ]; then
  echo "Expecting failIt to be empty."
  exit 1
fi

wantEmpty=$(gofmt -s -d -l . 2>&1 )
if [ -n "$wantEmpty" ]; then
  printf >&2 'gofmt failed for:\n%s\n' "$wantEmpty"
  failIt=1
fi

wantEmpty=$(goimports -l .)
if [ -n "$wantEmpty" ]; then
  printf >&2 'goimports failed for:\n%s\n' "$wantEmpty"
  failIt=1
fi

wantEmpty=$(go vet -all ./... 2>&1)
if [ -n "$wantEmpty" ]; then
  printf >&2 'govet failed for:\n%s\n' "$wantEmpty"
  failIt=1
fi

wantEmpty=$(golint ./...)
if [ -n "$wantEmpty" ]; then
  printf >&2 'golint failed for:\n%s\n' "$wantEmpty"
  failIt=1
fi

if [ -n "$failIt" ]; then
  unset failIt
  exit 1
fi

go test -v ./...
