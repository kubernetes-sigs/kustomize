#!/bin/bash

if [ -n "$failIt" ]; then
	echo "Expecting failIt to be empty."
  exit 1
fi

tmp=$(gofmt -s -d -l . 2>&1 )
if [ -n "$tmp" ]; then
		printf >&2 'gofmt failed for:\n%s\n' "$tmp"
		failIt=1
fi

tmp=$(goimports -l .)
if [ -n "$tmp" ]; then
		printf >&2 'goimports failed for:\n%s\n' "$tmp"
		failIt=1
fi

tmp=$(go vet -all ./... 2>&1)
if [ -n "$tmp" ]; then
		printf >&2 'govet failed for:\n%s\n' "$tmp"
		failIt=1
fi

tmp=$(golint ./...)
if [ -n "$tmp" ]; then
		printf >&2 'golint failed for:\n%s\n' "$tmp"
		failIt=1
fi

if [ -n "$failIt" ]; then
	unset failIt
  exit 1
fi

go test -v ./...
