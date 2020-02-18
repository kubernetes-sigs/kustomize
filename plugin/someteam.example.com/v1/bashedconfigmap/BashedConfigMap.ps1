Write-Output "
kind: ConfigMap
apiVersion: v1
metadata:
  name: example-configmap-test
  annotations:
    kustomize.config.k8s.io/needs-hash: `"true`"
data:
  username: `"$($args[1])`"
  password: `"$($args[2])`"
"
