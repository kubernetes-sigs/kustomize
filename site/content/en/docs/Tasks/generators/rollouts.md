---
title: "Managing Rollouts"
linkTitle: "Managing Rollouts"
weight: 3
date: 2023-10-06
description: >
  Rolling out ConfigMap and Secret updates
---

There are four different ways that you can use a ConfigMap to configure a container inside a Pod:

1. Inside a container command and args
2. Environment variables for a container
3. Add a file in read-only volume, for the application to read
4. Write code to run inside the Pod that uses the Kubernetes API to read a ConfigMap

It is important to note that rollout behavior depends on how the workload consumes the ConfigMap.
- ConfigMaps mounted in a volume are updated automatically.
- ConfigMaps consumed as environment variables are not updated automatically and require a pod restart.

It is common practice to perform a rolling update of the ConfigMap changes to Pods as soon as the ConfigMap changes are pushed.

Apply facilitates rolling updates for ConfigMaps by creating a new ConfigMap for each change to the data. Workloads (e.g. Deployments, StatefulSets, etc) are updated to point to a new ConfigMap instead of the old one. This allows the change to be gradually rolled the same way other Pod Template changes are rolled out.

Each generated Resources name has a suffix appended by hashing the contents. This approach ensures a new ConfigMap is generated each time the contents is modified.

**Note:** The Resource names will contain a suffix and will not match the exact name that is specified in the `kustomization.yaml` file. Consider this difference when looking for the Resource with `kubectl get`.
