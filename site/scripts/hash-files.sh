#!/bin/sh
# this script emits as hash for the files listed in $@
if command -v shasum >/dev/null 2>&1; then
  cat "$@" | shasum -a 256 | cut -d' ' -f1
elif command -v sha256sum >/dev/null 2>&1; then
  cat "$@" | sha256sum | cut -d' ' -f1
else
  echo "missing shasum tool" 1>&2
  exit 1
fi
