#!/bin/bash

echo "
kind: ConfigMap
apiVersion: v1
metadata:
  name: example-configmap-test
data:
  username: admin
  password: secret
"
