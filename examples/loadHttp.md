# load file from http

Resource and patch files could be loaded from http

<!-- @loadHttp -->
```sh
DEMO_HOME=$(mktemp -d)

cat <<EOF >$DEMO_HOME/kustomization.yaml
resources:
- https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/examples/helloWorld/configMap.yaml
EOF
```

<!-- @loadHttp -->
```sh
test 1 == \
  $(kustomize build $DEMO_HOME | grep "Good Morning!" | wc -l); \
  echo $?
```

Kustomize will try loading resource as a file either from local or http. If it
fails, try to load it as a directory or git repository.

Http load applies to patches as well. See full example in [loadHttp](loadHttp/).
