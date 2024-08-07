[SIG-CLI]: https://github.com/kubernetes/community/tree/master/sig-cli
[Slack channel]: https://kubernetes.slack.com/messages/kustomize
[Mailing list]: https://groups.google.com/forum/#!forum/kubernetes-sig-cli

[OWNERS file spec]: https://github.com/kubernetes/community/blob/master/contributors/guide/owners.md
[Kustomize OWNERS_ALIASES]: https://github.com/kubernetes-sigs/kustomize/blob/8049f7b1af52e8a7ec26faf6cf714f560d0043c5/OWNERS_ALIASES
[SIG-CLI Teams]: https://github.com/kubernetes/org/blob/main/config/kubernetes-sigs/sig-cli/teams.yaml
[Github permissions]: https://docs.github.com/en/organizations/managing-access-to-your-organizations-repositories/repository-permission-levels-for-an-organization#repository-access-for-each-permission-level

[Contributor License Agreement]: https://git.k8s.io/community/CLA.md
[Kubernetes Contributor Guide]: http://git.k8s.io/community/contributors/guide
[Contributor Cheat Sheet]: https://git.k8s.io/community/contributors/guide/contributor-cheatsheet/README.md
[CNCF Code of Conduct]: https://github.com/cncf/foundation/blob/master/code-of-conduct.md
[Kubernetes Community Membership]: https://github.com/kubernetes/community/blob/master/community-membership.md

[Kustomize Architecture]: ARCHITECTURE.md
[Contribution Guide]: https://kubectl.docs.kubernetes.io/contributing/kustomize/
[MacOS Dev Guide]: https://kubectl.docs.kubernetes.io/contributing/kustomize/mac/
[Windows Dev Guide]: https://kubectl.docs.kubernetes.io/contributing/kustomize/windows/

# Contributing Guidelines

Welcome to Kubernetes. We are excited about the prospect of you joining our [community](https://github.com/kubernetes/community)! The Kubernetes community abides by the [CNCF Code of Conduct]. Here is an excerpt:

_As contributors and maintainers of this project, and in the interest of fostering an open and welcoming community, we pledge to respect all people who contribute through reporting issues, posting feature requests, updating documentation, submitting pull requests or patches, and other activities._

## Getting Started

### Forking Kustomize and Working Locally
The Kustomize project uses a "Fork and Pull" workflow that is standard to GitHub. In git terms, your personal fork is referred to as the "origin" and the actual project's git repository is called "upstream". To keep your personal branch (origin) up to date with the project (upstream), it must be configured within your local working copy.

### Create a fork in GitHub
1. Visit https://github.com/kubernetes-sigs/kustomize
2. Click the `Fork` button on the top right

### Clone the repository
```bash
# Clone your repository fork from the previous step
git clone --recurse-submodules git@github.com:<your github username>/kustomize.git
cd kustomize

# Configure upstream
git remote add upstream https://github.com/kubernetes-sigs/kustomize
git remote set-url --push upstream no_push

# Review git configuration
git remote -v
```

### Create a working branch
```bash
# Fetch changes from upstream master
cd kustomize
git fetch upstream
git checkout master
git rebase upstream/master

# Create your working branch
git checkout -b myfeature
```

### Sync your working branch
You will need to periodically fetch changes from the `upstream` repository to keep your working branch in sync.
```bash
cd kustomize
git fetch upstream
git checkout myfeature
git rebase upstream/master
```

### Push to GitHub
When your changes are ready for review, push your working branch to your fork on GitHub.
```bash
cd kustomize
git push origin myfeature
```

### Pull Request Rules

We are using [Conventional Commits v1.0.0](https://www.conventionalcommits.org/en/v1.0.0/) as the main guideline of making PR. This guideline serves to help contributor and maintainer to classify their changes, thus providing better insight on type of release will be covered on each Kustomize release cycle.

1. Please add these keywords on your PR titles accordingly

| Keyword  | Description | Example |
| ------------- | ------------- | ------------- |
| fix  | Patching or fixing bugs or improvements introduction from previous release. This type of change will mark a `PATCH` release.  | fix: fix null value when generating yaml |
| feat  | New features. This change will mark a `MINOR` release. | feat: new transformer and generator for ACME API CRD. |
| chore  | Minor improvement outside main code base  | chore: add exclusion for transformer test. |
| ci | CI/CD related changes (e.g. github workflow, scripts, CI steps). | ci: remove blocking tests |
| docs  | Changes related to documentation. | docs: add rules documentation for PR. |


2. Add `BREAKING CHANGE:` on your commit message as footer to signify breaking changes. This will help maintainers identify `MAJOR` releases.
  
Example:

```
feat: change YAML parser from `yaml/v1` to `yaml/v2`

BREAKING CHANGE: parse() function now works with 2 arguments.
```

### Create a Pull Request

1. Visit your fork at `https://github.com/<user>/kustomize`
2. Click the **Compare & Pull Request** button next to your `myfeature` branch.
3. Check out the pull request [process](https://github.com/kubernetes/community/blob/master/contributors/guide/pull-requests.md) for more details and advice.

If you ran `git push` in the previous step, GitHub will return a useful link to create a Pull Request.

### Build Kustomize
The [Kustomize Architecture] document describes the respository organization and the kustomize build process.
```bash
# For go version >= 1.13
unset GOPATH
unset GO111MODULES

# Build kustomize binary and install in go bin path
cd kustomize
make kustomize

# Run unit tests
make test-unit-all

# Run linter
make lint

# Test examples against HEAD
make test-examples-kustomize-against-HEAD

# Run your development version
~/go/bin/kustomize version
```

### General resources for contributors

- [Contributor License Agreement] - Kubernetes projects require that you sign a Contributor License Agreement (CLA) before we can accept your pull requests.
- [Kubernetes Contributor Guide] - Main contributor documentation.
- [Contributor Cheat Sheet] - Common resources for existing developers.

Here are some additional ideas to help you get started with Kustomize:
- Attend a Kustomize Bug Scrub. Check the [SIG-CLI] meetings list to find the next one.
- Help triage issues by confirming validity and applying the appropriate `kind` label (e.g. comment `/kind bug`).
- Pick up an issue to fix. Issues with the `help-wanted` label are a good place to start, but you can also look for any issue with the `triage/accepted` label and no assignee. Remember to `/assign` yourself to let others know you're working on it.
- Help confirm new issues labelled `kind/bug` by reproducing them with the latest release.
- Support Kustomize users by responding to questions on issues labelled `kind/support` or in the [Slack channel].

## Mentorship

- [Mentoring Initiatives](https://git.k8s.io/community/mentoring) - We have a diverse set of mentorship programs available that are always looking for volunteers!

## Contributor Ladder

Kustomize follows the [Kubernetes Community Membership] contributor ladder. Roles are as follows:

1. Contributor: Anyone who actively contributes code, issues or reviews to the project. All contributors must sign the [Contributor License Agreement].
1. Reviewer: Contributors with a history of review and authorship on Kustomize. Has LGTM rights on the Kustomize repo (as do all kubernetes-sigs org members). Active contributors are encouraged to join the reviewers list to be automatically pinged on PRs.
1. Approver: Highly experienced active reviewer and contributor to Kustomize. Has both LTGM and approval rights on the Kustomize repo, as well as "maintain" [Github permissions].
1. Owner: Approver who sets technical direction and makes or approves design decisions for the project. Has LGTM and approval rights on the Kustomize repo as well as "admin" [Github permissions].

The kyaml module within the Kustomize repo has additional owners following the same ladder.

For the kustomize project, we have defined some specific guidelines on each step of the ladder:

To reach reviewer status, you must:
- Have been actively involved in kustomize for 3+ months
- Review at least 8 PRs that have been driven through to completion (see the reviewer guide below)
- Author at least 5 PRs that have been approved and merged
- Be a member of the kubernetes-sigs org. This should not be a blocker though, as once you meet the requirements for reviewer here,
the existing kustomize maintainers will be happy to sponsor your request to join the kubernetes-sigs org.
- Once you have met the above requirements, you may submit a PR adding yourself to the kustomize reviewers list, with links to your
contributions in the description.

To reach approver status, you must:
- Meet all the requirements of a reviewer
- Have been actively involved in kustomize for 6+ months
- Review at least 15 PRs that have been driven through to completion (see the reviewer guide below)
- Authored PRs meeting *either* of the following requirements:
  - 15 PRs that have been approved and merged
  - *OR* 10 PRs that have been approved and merged where some were more difficult, required greater thought/design,
or built up to larger features/long-term goals.
- File 3 issues. This can be any number of things, including but not limited to:
  - Bugs with kustomize usage that you've found
  - CI or release improvements
  - Creating subtasks of a larger feature or project that you are in charge of.
  - Long term improvements for the health of the project
- Triage at least 10 untriaged issues, including at least 1 feature request. The kustomize bug scrub is a great place to get practice with doing this, but you can
also follow the triage guide below to get started on your own.
- Demonstrate deeper understanding of kustomize goals. This can take many forms and is a bit subjective, but here are a few examples:
  - saying no to an eschewed feature, instead recommending an alternative solution that is more aligned with the declarative configuration model
  - active participation in discussion on a feature request issue
  - filing an issue describing a long term problem and solution aligned with kustomize goals, for example: https://github.com/kubernetes-sigs/kustomize/issues/5140
  - writing up KEPs for features that will improve the kustomize workflow while being aligned with kustomize goals, for example: https://github.com/kubernetes-sigs/kustomize/pull/4558
- Regularly interact with the existing kustomize maintainers, with clear communication about what you are working on or planning to work on. The kustomize
maintainers should know who you are and be familiar with your contributions.
- If you meet *most* of the above requirements while going above and beyond in a few areas, we will still consider your request to become an approver even
if you are missing one or two of the requirements. Please contact the maintainers directly to ask about getting approver status if you fall into this category.
- Otherwise, once you meet all the above requirements, you may:
  - request to be added to the kustomize maintainer meeting that occurs each week with the kustomize PMs.
  - submit a PR adding yourself to the kustomize approvers list, with links to your contributions in the description.

To reach owner status, you must:
- Meet all the requirements of an approver
- Have been actively involved with kustomize for 1+ year
- Assisted the current owner in driving the roadmap. This can be explicit or implicit help, such as:
  - Editing the roadmap directly
  - Reviewing the roadmap
  - Providing suggestions for issues or prioritization in meetings that indirectly influence the roadmap
- Regularly triage issues and attend the kustomize bug scrub
- Regularly review PRs (1-2 a week)
- Periodically lead the kustomize bug scrub
- Periodically release kustomize (ensuring that there are no release blockers and that release notes are clean)
- Be the primary owner or point of contact for a particular project or area of code
- Ideally, there should be 2-3 owners at a time. Reach out to the current owners if you are interested in ownership. These
requirements are not strict and evaluation is somewhat subjective.

## Reviewer guide
Please watch this talk on how to review code from Tim Hockin: https://www.youtube.com/watch?v=OZVv7-o8i40

For reviewing PRs in kustomize, we have some specific guidelines:
- If the PR is introducing a new feature:
  - *It must be implementing an issue that has already been triage/accepted or
a KEP that has been approved.* If it is not, then request the PR author to first file an issue.
  - The PR must include thorough tests for the new feature, including unit and integration tests
  - The code must be clean and readable, with thought given to how we will maintain the code in the future
  - If the feature requires being broken up into multiple PRs to ease review, the feature should not be exposed to users
    until the feature is completed in the last PR. For example, while we were building `kustomize localize`, we
    built the feature almost entirely under the `api` module as a library with all the needed tests. There was no way
    for users to invoke the localize code until the last PR that actually exposed the `kustomize localize` command in the
    kustomize binary. This allowed us to continue development of `kustomize localize` without blocking kustomize releases.
    If this type of development is not possible, then new features requiring multiple PRs should be
    developed in their own feature branch.
- If the PR is introducing a bug fix:
  - If the PR is not fixing an issue that has already been triage/accepted, follow the triage guide below on bug
    fixes to decide if this is a PR we want to accept.
  - The PR should have two distinct commits:
    - The first commit should add a test demonstrating incorrect behavior
    - The second commit should include the bug fix
    - Some sample PRs:
      - https://github.com/kubernetes-sigs/kustomize/pull/5263/commits
      - https://github.com/kubernetes-sigs/kustomize/pull/3931/commits
    - The regression test is absolutely required, and we cannot accept bug fixes without tests.
- If the PR is introducing a performance improvement:
  - The PR description should give an indication of how much the performance is being improved and how we
    can measure it - benchmark tests are fantastic.
- Other PRs (documentation, CI improvements, etc.) should be reviewed based on your best judgment.

## Triage guide
The possible triage labels are listed here: https://github.com/kubernetes-sigs/kustomize/labels?q=triage.

Triaging a feature request means:
- Understand what the user is asking for, and their use case.
- Verify that it is not an [eschewed feature](https://kubectl.docs.kubernetes.io/faq/kustomize/eschewedfeatures/#build-time-side-effects-from-cli-args-or-env-variables)
- Verify that it is not a duplicate issue.
- Look into workarounds. Is there another way that the user can achieve their use case with existing features?
- If you are new to this role, prior to leaving a comment on the issue, please bring it to weekly standup
for group discussion to make sure that we are all on the same page.
- Once you feel ready, you can label it with a triage label. Here's an [example](https://github.com/kubernetes-sigs/kustomize/issues/5049). You can also
look at other feature request issues to see how they were triaged and resolved. There are a few different triage labels that you can use, you can see the
full list [here](https://github.com/kubernetes-sigs/kustomize/labels?q=triage).

Triaging a bug means:
- First, verify that you can reproduce the issue. If you cannot reproduce the issue or need more information to give
it a go, triage it accordingly.
- Try to understand if this is really a bug or if this is intended behavior from kustomize. If it seems like intended
behavior, do your best to explain to the user why this is the case.
- If it seems to be a genuine bug, you can /triage accept the issue. In addition, investigate if there are workarounds or
alternative solutions for the user that they can try until the issue gets resolved.

The triage party for kustomize is here https://cli.triage.k8s.io/s/kustomize and can be a easy way to
find issues that have not been triaged yet.

## Project/Product Managers

Kustomize will have opportunities to join in a project/product manager role. You can reach out to
the existing kustomize maintainers if you are interested in this type of role. Project management work
can greatly help supplement your contributions as you climb from reviewer to approver
to owner.

Expectations for this role are:

- Triage 1 feature request each week, and bring it to weekly stand-up for discussion. Feature
requests are issues labeled kind/feature, and you can find them [here](https://github.com/kubernetes-sigs/kustomize/issues?q=is%3Aissue+is%3Aopen+kind+feature+label%3Akind%2Ffeature).
Please view the above triage guide for details on how to approach feature request triage.
- Monitor the kustomize Slack channel and try to help users if you can. It is a pretty
active channel, so responding to 4-5 users per week is sufficient even if some
questions go unanswered. If there is an interesting topic or a recurring problem that many
users are having, please bring it up in weekly stand-up.
- Keeping track of a queue of backlog issues or PRs that are not being actively looked at in any existing project board.
- Organizing or reorganizing project tracking boards when it makes sense.

You will also be asked to help with roadmap planning, deprecation communication, prioritization,
and doing research on kustomize usage when appropriate, though these responsibilities will occur less
frequently.

## Administrative notes:

- The [OWNERS file spec] is a useful resources in making changes.
- Maintainers and admins must be added to the appropriate lists in both [Kustomize OWNERS_ALIASES] and [SIG-CLI Teams]. If this isn't done, the individual in question will lack either PR approval rights (Kustomize list) or the appropriate Github repository permissions (community list).

## Contact Information

- [Slack channel]
- [Mailing list]
