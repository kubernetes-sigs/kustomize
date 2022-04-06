# Copyright 2022 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0

# Report how kustomize and the plugins use the API module.
#
# Usage:
#   ./hack/whatApi.sh plugin
#   ./hack/whatApi.sh kustomize

# The packages listed below in 'grep -v' lines will
# likely appear in API v1.
#
# Packages not listed (i.e. emitted to stdout) will
# likely move to internal.
#
function whatApi {
  echo "==== begin $1 =================================="
    find $1 -name "*.go" |\
    xargs grep \"sigs.k8s.io/kustomize/api/ |\
    sed 's|:\s"|: dummy "|'   |\
    sed 's|:\s\w\+\s"|  |'    |\
    sed 's|"$||'              |\
    awk '{ printf "%60s  %s\n", $2, $1 }' |\
    sed 's|sigs.k8s.io/kustomize/api/||' |\
    sort |\
    uniq |\
    grep -v " filesys "    |\
    grep -v " hasher "     |\
    grep -v " ifc "        |\
    grep -v " inventory "  |\
    grep -v " konfig"      |\
    grep -v " krusty "     |\
    grep -v " kv "         |\
    grep -v " provenance " |\
    grep -v " resid "      |\
    grep -v " resmap "     |\
    grep -v " resource "   |\
    grep -v " testutils"   |\
    grep -v " types "
  echo "==== end $1 =================================="
}


whatApi $1
