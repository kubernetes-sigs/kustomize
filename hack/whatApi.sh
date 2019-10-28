# Report on use of API module.
#
# Usage:
#   ./hack/whatApi.sh plugin
#   ./hack/whatApi.sh kustomize

function whatApi {
  echo "==== begin $1 =================================="
  find $1 -name "*.go" |\
    xargs grep \"sigs.k8s.io/kustomize/api/ |\
    sed 's|:\s"|: dummy "|'   |\
    sed 's|:\s\w\+\s"|  |'    |\
    sed 's|"$||'              |\
    awk '{ printf "%60s  %s\n", $2, $1 }' |\
    sed 's|sigs.k8s.io/kustomize/api/||' |\
    sort | uniq
  echo "==== end $1 =================================="
}

whatApi $1
