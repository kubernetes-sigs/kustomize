#!/bin/bash
# Runs pre-commit tests.
#
# Instead of failing on first error, complete all checks, then fail if need be.

# Echo go's version to verify against go's behavior below.
# Different versions do different things with respect to vendored code.
go version

# Assert state.
if [ -n "$failIt" ]; then
  echo "Expecting failIt to be empty."
  exit 1
fi

wantEmpty=$(gofmt -s -d -l . 2>&1 )
if [ -n "$wantEmpty" ]; then
  printf >&2 '\ngofmt failed for:\n%s\n' "$wantEmpty"
  failIt=1
fi

wantEmpty=$(goimports -l $(find . -type f -name '*.go' -not -path "./vendor/*") 2>&1)
if [ -n "$wantEmpty" ]; then
  printf >&2 '\ngoimports failed for:\n%s\n' "$wantEmpty"
  failIt=1
fi

wantEmpty=$(go vet -all ./... 2>&1)
if [ -n "$wantEmpty" ]; then
  printf >&2 '\ngo vet failed for:\n%s\n' "$wantEmpty"
  failIt=1
fi

wantEmpty=$(golint ./...)
if [ -n "$wantEmpty" ]; then
  printf >&2 '\ngolint failed for:\n%s\n' "$wantEmpty"
  failIt=1
fi

if [ -n "$failIt" ]; then
  unset failIt
  exit 1
fi

go test -v ./...

