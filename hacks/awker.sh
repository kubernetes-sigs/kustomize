function pluginDeps {
  find ./plugin -name "*.go" |\
    xargs grep sigs.k8s.io/kustomize/v3 |\
    sed 's|"sigs.k8s.io/kustomize/v3/||' |\
    awk '{ printf "%40s  %s\n", $2, $1 }' |\
    sed 's|"  \./| |' |\
    sed 's|:$||' |\
    sort | uniq
}


pluginDeps

