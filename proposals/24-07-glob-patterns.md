<!--
**Note:** When your proposal is complete, all of these comment blocks should be removed.

To get started with this template:

- [ ] **Make a copy of this file.**
  Name it `YY-MM-short-descriptive-title.md` (where `YY-MM` is the current year and month).
- [ ] **Fill out this file as best you can.**
  At minimum, you should fill in the "Summary" and "Motivation" sections.
- [ ] **Create a PR.**
  Ping `@kubernetes-sigs/kustomize-admins` and `@kubernetes-sigs/kustomize-maintainers`.
-->

# Add glob pattern support in the `resources` key of a user's `kustomization.yaml`

**Authors**:
- BitForger

**Reviewers**: <!-- List at least one Kustomize approver (https://github.com/kubernetes-sigs/kustomize/blob/master/OWNERS#L2) -->
- natasha41575
- KnVerey

**Status**: implementable
<!--
In general, all proposals made should be merged for the record, whether or not they are accepted.
Use the status field to record the results of the latest review:
- implementable: The default for this repo. If the proposal is merged, you can start working on it.
- deferred: The proposal may be accepted in the future, but it has been shelved for the time being.
A new PR must be opened to update the proposal and gain reviewer consensus before work can begin.
- withdrawn: The author changed their mind and no longer wants to pursue the proposal.
A new PR must be opened to update the proposal and gain reviewer consensus before work can begin.
- rejected: This proposal should not be implemented.
- replaced: If you submit a new proposal that supersedes an older one,
update the older one's status to "replaced by <link>".
-->

## Summary

<!--
In one short paragraph, summarize why this change is important to Kustomize users.
-->
File path globbing for the `resources` key enables users to maximize their use of Kustomize in their GitOps and CI/CD workflows. It simplifies their `kustomization.yaml` files by allowing them to pass a set of patterns that are equal to, or children of the `kustomization.yaml`'s working directory.

## Motivation

<!--
If this proposal is an expansion of an existing GitHub issue, link to it here.
-->
https://github.com/kubernetes-sigs/kustomize/issues/3205 is one of (at least two that I've seen) issues opened looking for this functionality.

As operational tasks shift to a declarative approach, the need to apply a large set of manifests from a "source of truth" repository is increasingly common. Currently, Kustomize very tediously requires a user to define every individual path for each resource the user wishes to build and apply. When you have multiple sets of services you need to apply, this can very quickly spiral into `kustomization.yaml` files with a 100 or more lines of resources. File-path globbing was a feature in the past, but was removed, much to the communities chagrin. This proposal seeks to reverse that change.

**Goals:**
<!--
List the specific goals of the proposal. What is it trying to achieve? How will we
know that this has succeeded?
-->
1. Resource resolution based on glob patterns
2. Resource resolution should only happen at the same or below the main `kustomization.yaml` file


**Non-goals:**
<!--
What is out of scope for this proposal? Listing non-goals helps to focus discussion
and make progress.
-->
1. Resource globbing should not accept exclude patterns

## Proposal

<!--
This is where we get down to the specifics of what the proposal actually is.
Include enough information to illustrate your proposal, but try not to
overwhelm reviewers with details. Focus on APIs and interfaces rather than implementation details,
e.g.:
- Does this proposal require new kinds, fields or CLI flags?
- Will this feature require extending the public interface of Kustomize's Go packages?
(it's ok if you're not sure yet)

A proof of concept PR is NOT required but is preferable to including large amounts of code
inline here, if you feel such implementation details are required to adequately explain your design.
If you have a PR, link to it at the top of this section.
-->
A feature flag will be added to the `kustomization.yaml` called `enableGlobbing` or `enableGlobSupport` that accepts a boolean value. When this flag is enabled, a user can pass glob patterns to the `resources` array and kustomize will take the pattern and collect all sibling and child paths that match the provided pattern.


### User Stories
<!--
Describe what people will be able to do if this KEP is implemented. If different user personas
will use the feature differently, consider writing separate stories for each.
Include as much detail as possible so that people can understand the "how" of the system.
The goal here is to make this feel real for users without getting bogged down.
-->

#### Story 1

Scenario summary: As a end user, I want to use a pattern to apply a bunch of manifests
<!--
A walkthrough of what it will look like for a user to take advantage of the new feature.
Include the steps the user will take and samples of the commands they'll run
and config they'll use.
-->

Take the below file structure

```
.
└── k8s/
    ├── pipelines/
    │   ├── pipeline_1/
    │   │   ├── pipeline.yaml
    │   │   ├── task.yaml
    │   │   └── nested/
    │   │       └── task.yaml
    │   └── pipeline_2/
    │       └── pipeline.yaml
    ├── tasks/
    │   ├── shared_task_1.yaml
    │   └── shared_task_2.yaml
    └── kustomization.yaml
```

A user could pass the following to kustomize to apply all the matching files

```yaml
resources:
  - pipelines/**/*.yaml
  - tasks/*/*.yaml
```

Kustomize would build all the matching files (using the already defined rules of kustomize, no name collisions, etc).

Glob support should work on files names, so a pattern like this should work

```
tasks/*/shared_task_*.yaml
```

#### Story 2

Scenario summary: As a end user, I want to apply a bunch of kustomizations using a pattern
<!--
A walkthrough of what it will look like for a user to take advantage of the new feature.
Include the steps the user will take and samples of the commands they'll run
and config they'll use.
-->

```
.
└── k8s/
    ├── bases/
    │   ├── app_1/
    │   │   ├── kustomization.yaml
    │   │   ├── deployment.yaml
    │   │   ├── service.yaml
    │   │   └── ...
    │   └── app_2/
    │       ├── kustomization.yaml
    │       ├── deployment.yaml
    │       ├── cronjob.yaml
    │       └── ...
    └── kustomization.yaml
```

A user could pass the following in their `kustomization.yaml`

```yaml
resources:
  - bases/**/
```

### Risks and Mitigations
<!--
What are the risks of this proposal, and how do we mitigate? Think broadly.
For example, consider both security, end-user privacy, and how this will
impact the larger Kubernetes ecosystem.
-->

There is the risk that a user could unintentionally include manifests if they choose to use the globbing support and they don't intentionally define their patterns to not be overly broad. For every other user, the file path globbing should be disabled by default so it shouldn't have an effect for them.

### Dependencies
<!--
Kustomize tightly controls its Go dependencies in order to remain approved for
integration into kubectl. It cannot depend directly on kubectl or apimachinery code.
Identify any new Go dependencies this proposal will require Kustomize to pull in.
If any of them are large, is there another option?
-->

Go has a built in filepath package that should be able to support the described use cases.

### Scalability
<!--
Is this feature expected to have a performance impact?
Explain to what extent and under what conditions.
-->

There may be a performance impact by using glob searching over looking up each individually named resource. These may scale with the amount of globs to lookup. I would have to do some testing to see how that behaved.

## Drawbacks
<!--
Why should this proposal _not_ be implemented?
-->



## Alternatives
<!--
What other approaches did you consider, and why did you rule them out? Be concise,
but do include enough information to express the idea and why it was not acceptable.
-->

Currently, the only other approach is to use a script to re-generate the `kustomization.yaml` every time you go to apply a lot of manifests. This is unnecessary complexity for something many other ops tools support.

## Rollout Plan
<!--
Depending on the scope of the features and the risks enabling it implies,
you may need to use a formal graduation process. If you don't think this is
necessary, explain why here, and delete the alpha/beta/GA headings below.
-->

I don't believe this will need a formal graduation process as the proposal adds the feature behind an option flag in the `kustomization.yaml`. The defualt behavior of the `resources` key would not change.

<!-- ### Alpha -->
<!--
New Kinds should be introduced with an alpha group version.
New major features should often be gated by an alpha flag at first.
New transformers can be introduced for use in the generators/validators/transformers fields
before they get their own top-level field in Kustomization.
-->

- Will the feature be gated by an "alpha" flag? Which one?
- Will the feature be available in `kubectl kustomize` during alpha? Why or why not?

<!-- ### Beta -->
<!--
If the alpha was not available in `kubectl kustomize`, you need a beta phase where it is.
Full parity with `kubectl kustomize` is required at this stage.
-->

<!-- ### GA -->
<!--
You should generally wait at least two `kubectl` release cycles before promotion to GA,
to ensure that the broader user base has time to try the feature and provide feedback.
For example, if your feature first appears in kubectl 1.23, promote it in 1.25 or later.
-->
