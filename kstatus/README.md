# kstatus

kstatus provides tools for checking the status of Kubernetes resources. The primary use case is knowing when 
(or if) a given set of resources in cluster has successfully reconciled an apply operation.

## Concepts

This effort has several goals, some with a shorter timeline than others. Initially, we want to provide
a library that makes it easier to decide when changes to a set of resources have been reconciled in a cluster.
To support types that do not yet publish status information, we will initially fallback on type specific rules. 
The library already contains rules for many of the most common built-in types such a Deployment and StatefulSet.

For custom resource definitions (CRDs), there currently isn't much guidance on which properties should be exposed in the status
object and which conditions should be used. As part of this effort we want to define a set of standard conditions
that the library will understand and that we encourage developers to adopt in their CRDs. These standard conditions will
be focused on providing the necessary information for understanding status of the reconcile after `apply` and it is not
expected that these will necessarily be the only conditions exposed in a custom resource. Developers will be free to add as many conditions
as they wish, but if the CRDs adopt the standard conditions defined here, this library will handle them correctly.

The `status` objects for built-in types don't all conform to a common behavior. Not all built-in types expose conditions,
and even among the types that does, the types of conditions vary widely. Long-term, we hope to add support for the
standard conditions to the built-in types as well. This would remove the need for type-specific rules for determining
status.

### Statuses

The library currently defines the following statuses for resource:
* __InProgress__: The actual state of the resource has not yet reached the desired state as specified in the
resource manifest, i.e. the resource reconcile has not yet completed. Newly created resources will usually 
start with this status, although some resources like ConfigMaps are Current immediately.
* __Failed__: The process of reconciling the actual state with the desired state has encountered and error
or it has made insufficient progress.
* __Current__: The actual state of the resource matches the desired state. The reconcile process is considered
complete until there are changes to either the desired or the actual state.
* __Terminating__: The resource is in the process of being deleted.
* __Unknown__: This is for situations when the library are unable to determine the status of a resource.

### Conditions

The conditions defined in the library are designed to adhere to the "abnormal-true" pattern, i.e. that 
conditions are present and with a value of true whenever something unusual happens. So the absence of
any conditions means everything is normal. Normal in this situation simply means that the latest observed 
generation of the resource manifest by the controller have been fully reconciled with the actual state. 

* __InProgress__: The controller is currently working on reconciling the latest changes.
* __Failed__: The controller has encountered an error during the reconcile process or it has made
insufficient progress (timeout).

The use of the "abnormal-true" pattern has some challenges. If the controller is not running, or for some
reason not able to update the resource, it will look like it is in a good state when that is not true. The
solution to this issue is to adopt the pattern used by several of the built-in types where there is an
`observedGeneration` property on the status object which is set by the controller during the reconcile loop.
If the `generation` and the `observedGeneration` of a resource does not match, it means there are changes
that the controller has not yet seen, and therefore not acted upon.

## Features

The library is currently separated into two packages, one that provides the basic functionality, and another that
builds upon the basics to provide a higher level API.

**sigs.k8s.io/kustomize/kstatus/status**: Provides two basic functions. First, it provides the `Compute` function
that takes a single resource and computes the status for this resource based on the fields in the status object for
the resource. Second, it provides the `Augment` function that computes the appropriate standard conditions based on
the status object and then amends them to the conditions in the resource. Both of these functions currently operate
on Unstructured types, but this should eventually be changed to rely on the kyaml library. Both of these functions 
compute the status and conditions solely based on the data in the resource passed in. It does not communicate with
a cluster to get the latest state of the resources.

**sigs.k8s.io/kustomize/kstatus/wait**: This package builds upon the status package and provides functionality that
will fetch the latest state from a cluster. It provides the `FetchAndResolve` function that takes list of resource
identifiers, fetches the latest state for all the resources from the cluster, computes the status for all of them and
returns the results. `WaitForStatus` accepts a list of resource identifiers and will poll cluster for the status of 
the resources until all resources have reached the `Current` status.

## Challenges

### Status is not obvious for all resource types

For some types of resources, it is pretty clear what the different statuses mean. For others, it
is far less obvious. For example, what does it mean that a PodDisruptionBudget is Current? Based on
the assumptions above it probably should be whenever the controller has observed the resource
and updated the status object of the PDB with information on allowed disruptions. But currently, a PDB is
considered Current when the number of healthy replicas meets the threshold given in the PDB. Also, should
the presence of a PDB influence when a Deployment is considered Current? This would mean that a Deployment
should be considered Current whenever the number of replicas reach the threshold set by the corresponding
PDB. This is not currently supported as described below.

### Status is decided based on single resource
Currently the status of a resource is decided solely based on information from
the state of that resource. This is an issue for resources that create other resources
and that doesn't provide sufficient information within their own status object. An example
is the Service resource that doesn't provide much status information but do generate Endpoint
resources that could be used to determine status. Similar, the status of a Deployment could be 
based on its generated ReplicaSets and Pods.

Not having the generated resources also limits the amount of details that can be provided
when something isn't working as expected.


## Future

### Depend on kyaml instead of k8s libraries
The sigs.k8s.io/kustomize/kstatus/status package currently depends on k8s libraries. This can be 
challenging if someone wants to vendor the library within their own project. We want to replace
the dependencies on k8s libraries with kyaml for the status package. The wait package needs to 
talk to a k8s cluster, so this package will continue to rely on the k8s libraries.

### Use watches instead of polling

We currently poll for updates to resources, but it would be possible to set up
watches instead. This could also be combined with deciding status based on not only a single
resource, but also all its generated resources. This would lead to a design that seems similar
to a controller, so maybe a solution like this could be built on top of controller-runtime. 
A challenge here is that the rules for each built-in type would need to be expressed in a different
way that what we currently do.