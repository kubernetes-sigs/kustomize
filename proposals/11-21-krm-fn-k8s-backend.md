# Implementation of k8s backend to run krm functions

**Authors**: alexey.odinokov.82@gmail.com

**Reviewers**: <!-- List at least one Kustomize approver (https://github.com/kubernetes-sigs/kustomize/blob/master/OWNERS#L2) -->

**Status**: implementable

## Summary

Currently Kustomize supports plugin-extensions via containers that use the certain API:
see [KRM-functions](https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md).
This is very convenient because of many reasons, e.g. it doesn't require prior plugin-installation, and in the current
implementation it requires only `docker` installed and running: `kustomize` calls `docker` client
to run the specified docker-container. Unfortunately, in cases when `kustomize` has to be ran
inside another container, e.g. inside a pod like in case of k8s it requires to use docker-in-docker or
mount docker-socket, which is not always possible. And in cases when it's possible this in its turn
creates security concerns, because docker-in-docker requires container to be ran in privileged mode
and shared docker-socket also introduces issues: it allows to control containers created by others. 
In addition, its `--mount` flag won't work correctly, because source path is treated like a path in the
host environment, not inside the container. And that allows to expose some files from host to the container.

This proposal aims to avoid that concerns by introducing the alternative approach where krm-functions
are ran as a pod in k8s cluster. That approach allows to use k8s-native mechanisms to control that pod.
Moreover, since `kustomize` is included into `kubectl` that approach removes a mandatory dependency
on `docker` and makes `kubectl` self-sufficient to run krm-function based plugins.

## Motivation

1. https://github.com/GoogleContainerTools/kpt/issues/2158
1. https://github.com/kubernetes-sigs/kustomize/pull/4260


**Goals:**

1. To enrich container-based plugin-extension implementation by introducing the new backend that
will be an alternative to `docker` and that will allow to run krm-functions inside k8s using its native mechanisms to run containers
1. To use k8s-native mechanisms to control access of krm-function to the environment it is running in


**Non-goals:**

1. Change Krm-functions spec
1. Add new fields/settings to container-based kustomize plugins-extensions

## Proposal

The basis of this proposal is to give ability for the user to select that the krm-function would be ran
*as-a-pod* rather than using `docker` (which will still be a default behavior that will guarantee no braking changes).
For that purpose, in addition to the existing `--enable-exec`, `--enable-starlark` flags that have a very similar
purpose - to allow the user *explicitly* say that it's ok to run krm-function not only in docker, we propose to add
`--enable-as-a-pod` flag that will switch backend from `docker` to `kubectl`.

In case the user has several k8s clusters, it's necessary to give him the ability to select
where the krm-container will run. Kubectl has a pretty big list of parameters that allow
to set kubeconfig location, context, namespace and etc. Adding all of them would be troublesome, but
since we're only passing all of them to kubectl it's much easier to introduce only one parameter,
`--pod-kubectl-args` that will contain all that arguments as one string, e.g.
`--context kind-kustomize-api-test -n special-namespace`. This will give the user the maximum
flexibility without limiting him. Even if kubectl will add some new parameters in future they all automatically
will be supported.

The current `docker` based implementation runs container in as-hermetic-as-possible environment:

1. only env variables specified in function config OR env variables explicitly listed by the user are passed to krm-container.
1. if krm-function requires network access it can be enabled only by explicitly set `--network` flag.
1. if krm-function needs some additional mounted file/directory this also can be done only via setting a special `--mount` flag explicitly.

That means that all non-default non-hermetic features except env variables are restricted and can be set only explicitly.
Similarly, the restrictions applied to the k8s pod has to be set *explicitly* by the user with the same exception of env variables.

Unfortunately, it's not possible to completely block network for a pod - networking is a k8s crucial part. But if needed
it's possible to apply NetworkPolicies to the pod that match to it by labels. Since NetworkPolicy is a standard method
to restrict pods, the user can apply one or more of them to the pods that run krm-function.
NetworkPolicies can be applied only to the pods that match the label or expression that consists of set of labels.
That means we have to give ability to user to set an arbitrary label(s) for the created pod.

To make pod-based implementation compatible with major number of already existing set of krm-functions
it's necessary for some of them to allow mounted files/directories if the user explicitly sets that.
In contrast to the 'docker'-backend use-cases where user shares with krm a local file or directory,
k8s doesn't allow to do this. That means that some krm-functions that require mounted files/directories
just won't work with `--enable-as-a-pod` when they run from the client host. BUT if kustomize is ran from
another pod that uses k8s volume, it will be possible to re-share the same volume that kustomize is using to
the pod where krm-function is running. That means we have to give ability to user to configure a set of volumes for the pod.

Keeping all that in mind now we can speak about the implementation options:

The first option would be to give ability to user to simply set pod-label and list of volumes.
This could be resolved by introducing of one more extra argument for labels and in reusing of the existing `--mount` argument.
Or even with only one `--mount` argument, if we agree that the label will be always the constant `app: krm-pod`.
We could select some default template for Pod we create, put there env vars, volumes and set the label and that implementation would work.

Another approach is to allow the user to use the maximum set of possible options that k8s allows when Pod is created, e.g.,
in addition to the existing volumes passed with arguments, the user should be able to provide ConfigMaps that also have to be
mounted. That approach leads us to ability to create PodTemplate in advance and to use it to build krm-function pod.
It will have a very similar outcome from security perspective with default values, but since it allows the user
much more flexibility doing things in k8s-native way, this approach is selected to be implemented.
It is implemented in [this PR](https://github.com/kubernetes-sigs/kustomize/pull/4260), but the current implementation
doesn't allow to set the pod label via PodTemplate(TBD). To be able to do so we can either copy labels from PodTemplate, always set the as
constant fields, or introduce one more extra argument. Since it's desired to minimize the number of arguments, the first option
with copying labels from template will be selected. User may set the PodTemplate name with `--pod-template-name` argument and the pod
template must exist in the same namespace where krm-function pod is going to be ran. The user may skip that parameter and by
default the [following PodTemplate](https://github.com/kubernetes-sigs/kustomize/pull/4260/files#diff-bcc3694e89726203de4e763f2130891ec669833ab26f3856bcddde18c6514a9fR27)
will be used. Env variables are put to that template already (TBD: maybe it's better to append env vars, rather than just replace the env vars already presented in the
template). `--mount` flag is ignored as of now in case kubectl-backend is used, but there is a possible option to take what the set of volumes from PodTemplate
and to append that with data from `--mount` flag - this is a potential topic for discussion.

There is one more potential improvement that should be taken into account: with docker backend krm-function couldn't run
another krm-functions. For that reason kustomize krm-function was deleted - kustomize couldn't run
krm-functions without docker-in-docker. With kubectl backend if we could specify service account for the pod
running krm-function we could run another krm-function initiated by the original krm-function and our
example with kustomize-krm-function could become possible and security-aspects could be controlled with
standard k8s-native configuration in the PodTemplate + couple of additional k8s resources.
This is one more benefit of the approach with `--pod-template-name` flag.

The proposed implementation has one more flag `--pod-start-timeout`. It was added, because depending on the cluster and the image size to pull the
time for pod start may vary. If the user doesn't set that, the timeout is 60s by default. The previous implementation of RunFn in kyaml
didn't work with timeouts. In contrast the updated modification that is used in `kpt` has timeouts that are applicable to the whole container execution including
time needed to pull the image. In general, it's a good practice for automated tools to be able to set timeouts for operations. One potential topic
for discussion may be - to create a common approach for that, e.g., to set only the overall timeout E.g. if timeout for that particular krm-functions is set
for the whole execution, this timeout can be set in function config and we may get rid of this argument. That means we may want to stick to the approach
that `kpt` selected.

We're using the following sequence to run krm-function in the pod:
1. If PodTemplateName is set - taking it from k8s by kubectl (the same namespace), OR using defaultPodTemplate overwise. See some mandatory fields values there, e.g. stdin,
 that have to be set in the PodTemplate to work correctly with 'kubectl' backend.
1. Creating pod using PodTemplate spec and applying imageUrl and set of envs
1. Waiting maximum PodStartTimeout for the pod up and running
1. using kubectl attach -q to send input and to get output
1. deleting the created pod (deferred)

One of the possible alternatives was to implement kubectl backend via k8s API instead of just simple calls to kubectl binary.
There are pros and cons of each approach.
The benefits of API-approach could be to avoid prerequisite of `kubectl` installed.
The drawbacks are:
1. A bit more complicated implementation
1. I wasn't sure that this won't create issues in the process of integrating new version of kustomize into kubectl. if this is a real issue, kustomize shouldn't try to implement
that backend via api. But `kpt` may want this option, because it's cleaner.
1. No need to implement the same args as kubectl has that we pass to `--pod-kubectl-args`
1. May be more, but these were enough to not to go this path

In the current implementation version, the whole content of `container` folder was moved to `docker` folder and that package was renamed.
The `kubectl`-based backend is located in `kubectl` folder.
The new `container` package was created that aggregated docker and kubectl that receives all parameters from filter-provider located in RunFn module.

The unit-tests of the PR rely on kind k8s cluster that is started BEFORE api unit-tests and stopped AFTER. As of now this behavior was added
into the Makefile. Maybe it's better to create a special scripts in some place and call them from the `fnplugin_test.go` before and after tests. - one more topic
to discuss.


### User Stories

#### Story 1

Scenario summary: As an end user that has a working k8s cluster I want to run kustomizations with krm-function plugin in that cluster

If the user has a working kubeconfig and kubectl installed he may follow the next steps:
1. Select the appropriate namespace for running krm-functions (may require creation of one)
1. Optionally create PodTemplate in the selected namespace
1. Run kustomize as ususal but with extra flag `--enable-as-a-pod --pod-kubectl-args "-n <selected namespace> <additional kubectl flags if needed>"` + optionally with `--pod-template-name <created PodTeamplateName>`
1. Krm-function must be executed in the cluster and in the selected namespace. There is no need in installing docker locally

#### Story 2

Scenario summary: As a CICD engineer, I want to run kustomizations that includes krm-functions plugins in the pod

In order to run such pod (that may be a part of some CICD pipeline, e.g. tekton pipeline) it will be necessary to:
1. Select the appropriate namespace for running krm-functions (may require creation of one) prior to running everything
1. Optionally create PodTemplate in the selected namespace
1. Create a service account for the pod there kustomize will live and allow it to create pods in the selected namespace
1. Optionally if it's required to have a shared volume between kustomize pod and krm-function pod - create that and
configure kustomize pod and krm-function PodTemplate appropriately
1. Run the kustomize pod and make sure that it kustomize is ran with `--enable-as-a-pod` + required additional flags as it's specified in Story 1.

### Risks and Mitigations

In general, the apporach works already and there are no serious risks

### Dependencies

`kubectl` installed on the host/in the container

### Scalability

This backend seems like starts pods a bit slower than docker (at least in the current implementation).
But it shouldn't be a very serious issue now.

## Drawbacks

n/a

## Alternatives

Different alternatives on what controls the user should have are described in the `Proposal` section.

## Rollout Plan

This feature is compatible with `docker`-based implementation and reuses the same function config.
plugin configuration will stay the same, the only difference will be in the ability to run plugins on top of k8s.

### Alpha

- Will the feature be gated by an "alpha" flag? Which one?

all external plugins are still considered as alpha-plugins

- Will the feature be available in `kubectl kustomize` during alpha? Why or why not?

yes, it should be available, because in case of kubectl only kubectl binary will be necessary to
run everything.

### Beta

n/a

### GA

n/a
