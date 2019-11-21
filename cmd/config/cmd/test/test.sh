#!/bin/bash
# Copyright 2019 The Kubernetes Authors.
# SPDX-License-Identifier: Apache-2.0




cat <<End-of-message
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${NAME}
  labels:
    app: nginx
    name: ${NAME}
spec:
  replicas: ${REPLICAS}
  selector:
    matchLabels:
      app: nginx
      name: ${NAME}
  template:
    metadata:
      labels:
        app: nginx
        name: ${NAME}
    spec:
      containers:
      - name: ${NAME}
        image: nginx:v1.7
        ports:
        - containerPort: 8080
          name: http
---
apiVersion: v1
kind: Service
metadata:
  name: ${NAME}
  labels:
    app: nginx
    name: ${NAME}
spec:
  ports:
  # This i the port.
  - port: 8080
    targetPort: 8080
    name: http
  selector:
    app: nginx
    name: ${NAME}
End-of-message
