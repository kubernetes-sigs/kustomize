// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package kstatus contains libraries for computing status of kubernetes
// resources.
//
// Status
// Get status and/or conditions for resources based on resources already
// read from a cluster, i.e. it will not fetch resources from
// a cluster.
//
// Wait
// Get status and/or conditions for resources by fetching them
// from a cluster. This supports specifying a set of resources as
// an Inventory or as a list of manifests/unstructureds. This also
// supports polling the state of resources until they all reach a
// specific status. A common use case for this can be to wait for
// a set of resources to all finish reconciling after an apply.
package kstatus
