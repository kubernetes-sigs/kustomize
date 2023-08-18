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

[Contribution Guide]: https://kubectl.docs.kubernetes.io/contributing/kustomize/
[MacOS Dev Guide]: https://kubectl.docs.kubernetes.io/contributing/kustomize/mac/
[Windows Dev Guide]: https://kubectl.docs.kubernetes.io/contributing/kustomize/windows/

# Contributing Guidelines

Welcome to Kubernetes. We are excited about the prospect of you joining our [community](https://github.com/kubernetes/community)! The Kubernetes community abides by the [CNCF Code of Conduct]. Here is an excerpt:

_As contributors and maintainers of this project, and in the interest of fostering an open and welcoming community, we pledge to respect all people who contribute through reporting issues, posting feature requests, updating documentation, submitting pull requests or patches, and other activities._

## Getting Started

Dev guides:

- [Contribution Guide]
- [MacOS Dev Guide]
- [Windows Dev Guide]

General resources for contributors:

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

Administrative notes:

- The [OWNERS file spec] is a useful resources in making changes.
- Maintainers and admins must be added to the appropriate lists in both [Kustomize OWNERS_ALIASES] and [SIG-CLI Teams]. If this isn't done, the individual in question will lack either PR approval rights (Kustomize list) or the appropriate Github repository permissions (community list).

## Project/Product Managers

Kustomize will have opportunities to join in a project/product manager role. You can
typically start working on this role as part of a kustomize training cohort, so please
keep an eye out for that or reach out to the leads if you are interested in this type of
work. Expectations for this role are:

- Triage 1 feature request each week, and bring it to weekly stand-up for discussion. Feature
requests are issues labeled kind/feature, and you can find them [here](https://github.com/kubernetes-sigs/kustomize/issues?q=is%3Aissue+is%3Aopen+kind+feature+label%3Akind%2Ffeature).
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
- Monitor the kustomize Slack channel and try to help users if you can. It is a pretty
active channel, so responding to 4-5 users per week is sufficient even if some
questions go unanswered. If there is an interesting topic or a recurring problem that many
users are having, please bring it up in weekly stand-up.
- Keeping track of a queue of backlog issues or PRs that are not being actively looked at in any existing project board.
- Organizing or reorganizing project tracking boards when it makes sense.

You will also be asked to help with roadmap planning, deprecation communication, prioritization,
and doing research on kustomize usage when appropriate, though these responsibilities will occur less
frequently.

## Contact Information

- [Slack channel]
- [Mailing list]
