// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package wait contains functionality for getting the statuses
// of a list of kubernetes resources. Unlike the status package,
// the functions exposed in the wait package will talk to a
// live kubernetes cluster to get the latest state of resources
// and provides functionality for polling the cluster until the
// resources reach the Current status.
//
// FetchAndResolve will fetch resources from a cluster, compute the
// status for each of them and then return the results. The list of
// resources is defined as a slice of ResourceIdentifier, which is
// an interface that is implemented by the Unstructured type. It
// only requires functions for getting the apiVersion, kind, name
// and namespace of a resource.
//
//   import (
//     "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
//     "k8s.io/apimachinery/pkg/types"
//     "sigs.k8s.io/kustomize/kstatus/wait"
//   )
//
//   key := types.NamespacedName{Name: "name", Namespace: "namespace"}
//   deployment := &unstructured.Unstructured{
//     Object: map[string]interface{}{
//       "apiVersion": "apps/v1",
//       "kind":       "Deployment",
//     },
//   }
//   client.Get(context.Background(), key, deployment)
//   resourceIdentifiers := []wait.ResourceIdentifier{deployment}
//
//   resolver := wait.NewResolver(client)
//   results := resolver.FetchAndResolve(context.Background(), resourceIdentifiers)
//
// WaitForStatus also looks up status for a list of resources, but it will
// block until all the provided resources has reached the Current status or
// the wait is cancelled through the passed-in context. The function returns
// a channel that will provide updates as the status of the different
// resources change.
//
//   import (
//     "sigs.k8s.io/kustomize/kstatus/wait"
//   )
//
//   resolver := wait.NewResolver(client)
//   eventsChan := resolver.WaitForStatus(context.Background(), resourceIdentifiers, 2 * time.Second)
//   for {
//     select {
//     case event, ok := <-eventsChan:
//       if !ok {
//         return
//       }
//       fmt.Printf(event) // do something useful here.
//     }
//   }
package wait
