/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package defaultconfig

const commonLabelFieldSpecs = `
commonLabels:
- path: metadata/labels
  create: true

- path: spec/selector
  create: true
  version: v1
  kind: Service

- path: spec/template/metadata/labels
  create: true
  version: v1
  kind: ReplicationController

- path: spec/template/metadata/labels
  create: true
  kind: Deployment

- path: spec/template/spec/affinity/podAffinity/preferredDuringSchedulingIgnoredDuringExecution/podAffinityTerm/labelSelector/matchLabels
  create: false
  group: apps
  kind: Deployment


- path: spec/template/spec/affinity/podAntiAffinity/preferredDuringSchedulingIgnoredDuringExecution/podAffinityTerm/labelSelector/matchLabels
  create: false
  group: apps
  kind: Deployment

- path: spec/template/metadata/labels
  create: true
  kind: ReplicaSet

- path: spec/template/metadata/labels
  create: true
  kind: DaemonSet

- path: spec/template/metadata/labels
  create: true
  group: apps
  kind: StatefulSet

- path: spec/template/spec/affinity/podAffinity/preferredDuringSchedulingIgnoredDuringExecution/podAffinityTerm/labelSelector/matchLabels
  create: false
  group: apps
  kind: StatefulSet

- path: spec/template/spec/affinity/podAntiAffinity/preferredDuringSchedulingIgnoredDuringExecution/podAffinityTerm/labelSelector/matchLabels
  create: false
  group: apps
  kind: StatefulSet

- path: spec/volumeClaimTemplates[]/metadata/labels
  create: true
  group: apps
  kind: StatefulSet

- path: spec/template/metadata/labels
  create: true
  group: batch
  kind: Job

- path: spec/jobTemplate/metadata/labels
  create: true
  group: batch
  kind: CronJob

- path: spec/jobTemplate/spec/template/metadata/labels
  create: true
  group: batch
  kind: CronJob
`
