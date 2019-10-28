# Usage:
#   ./hack/showDeps ./plugin
#   ./hack/showDeps ./kustomize

function showDeps {
  echo "==== begin $1 =================================="
  find $1 -name "*.go" |\
    xargs grep \"sigs.k8s.io/kustomize/api/ |\
		sed 's|:\s"| |' |\
    sed 's|"$||' |\
    awk '{ printf "%60s  %s\n", $2, $1 }' |\
		sed 's|sigs.k8s.io/kustomize/api/||' |\
    sort | uniq
  echo "==== end $1 =================================="
}

showDeps $1
