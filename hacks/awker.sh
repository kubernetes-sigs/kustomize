function showDeps {
  echo "==== begin $1 =================================="
  find $1 -name "*.go" |\
    xargs grep \"sigs.k8s.io/kustomize/??? |\
    grep -v "/api/" |\
    sed 's|"sigs.k8s.io/kustomize/v3/||' |\
    awk '{ printf "%40s  %s\n", $2, $1 }' |\
    sed 's|"  \./| |' |\
    sed 's|:$||' |\
    sort | uniq
  echo "==== end $1 =================================="
}


showDeps ./plugin
showDeps ./kustomize
