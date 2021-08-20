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

# Your short, descriptive title

**Authors**:

**Reviewers**:

**Status**: implementable
<!--
In general, all proposals made should be merged for the record, whether or not they are accepted. Use the status field to record the results of the latest review:
- implementable: The default for this repo. If the proposal is merged, you can start working on it.
- deferred: The proposal may be accepted in the future, but it has been shelved for the time being. A new PR must be opened to update the proposal and gain reviewer consensus before work can begin.
- withdrawn: The author changed their mind and no longer wants to pursue the proposal. A new PR must be opened to update the proposal and gain reviewer consensus before work can begin.
- rejected: This proposal should not be implemented.
- replaced: If you submit a new proposal that supersedes an older one, update the older one's status to "replaced by <link>".
-->

## Summary

<!--
This section is incredibly important for producing high-quality, user-focused
documentation such as release notes or a development roadmap.
A good summary is probably at least a paragraph in length.
-->

## Motivation

<!--
This section is for explicitly listing the motivation, goals, and non-goals of
this proposal.  Describe why the change is important and the benefits to users. If this
proposal is an expansion of an existing GitHub issue, link to it here.
-->

### Goals

<!--
List the specific goals of the proposal. What is it trying to achieve? How will we
know that this has succeeded?
-->
1. A goal
1. Another goal

### Non-Goals

<!--
What is out of scope for this proposal? Listing non-goals helps to focus discussion
and make progress.
-->
1. A non-goal
1. Another non-goal

## Proposal

Proof of concept PR (optional):

<!--
This is where we get down to the specifics of what the proposal actually is. What is the plan for implementing this feature? Does this proposal require new kinds, fields or flags? What do they look like?
-->

### User Stories
<!--
Describe what people will be able to do if this KEP is implemented. If different user personas
will use the feature differently, consider writing separate stories for each.
Include as much detail as possible so that people can understand the "how" of the system.
The goal here is to make this feel real for users without getting bogged down.
-->

#### Story 1

Scenario summary: As a [end user, extension developer, ...], I want to [...]
<!--
A walkthrough of what it will look like for a user to take advantage of the new feature.
Include the the steps the user will take and samples of the commands they'll run
and config they'll use.
-->

#### Story 2

Scenario summary: As a [end user, extension developer, ...], I want to [...]
<!--
A walkthrough of what it will look like for a user to take advantage of the new feature.
Include the the steps the user will take and samples of the commands they'll run
and config they'll use.
-->

### Risks and Mitigations
<!--
What are the risks of this proposal, and how do we mitigate? Think broadly.
For example, consider both security and how this will impact the larger
Kubernetes ecosystem.
-->

### Dependencies
<!--
Kustomize tightly controls its Go dependencies in order to remain approved for
integration into kubectl. It cannot depend directly on kubectl or apimachinery code.
Identify any new Go dependencies this proposal will require Kustomize to pull in.
If any of them are large, is there another option?
-->

### Scalability
<!--
Is this feature expected to have a performance impact?
Explain to what extent and under what conditions.
-->

## Drawbacks
<!--
Why should this proposal _not_ be implemented?
-->

## Alternatives
<!--
What other approaches did you consider, and why did you rule them out? Be concise,
but do include enough information to express the idea and why it was not acceptable.
-->

## Rollout Plan
<!--
Depending on the scope of the features and the risks enabling it implies,
you may need to use a formal graduation process. If you don't think this is
necessary, explain why here, and delete the alpha/beta/GA headings below.
-->

### Alpha
<!--
New Kinds should be introduced with an alpha group version.
New major features should often be gated by an alpha flag at first.
New transformers can be introduced for use in the generators/validators/transformers fields
before they get their own top-level field in Kustomization.
-->

- Will the feature be gated by an "alpha" flag? Which one?
- Will the feature be available in `kubectl kustomize` during alpha? Why or why not?

### Beta
<!--
If the alpha was not available in `kubectl kustomize`, you need a beta phase where it is.
Full parity with `kubectl kustomize` is required at this stage.
-->

### GA
<!--
You should generally wait at least two `kubectl` release cycles before promotion to GA,
to ensure that the broader user base has time to try the feature and provide feedback.
For example, if your feature first appears in kubectl 1.23, promote it in 1.25 or later.
-->
