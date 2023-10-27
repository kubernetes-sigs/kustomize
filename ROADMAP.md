# Kustomize roadmap 2023-2024

This document describes the items that we hope to make progress on over the next
1 year (H2 2023 and H1 2024). Take this roadmap as more of what we hope to achieve
rather than what we promise to achieve, as some items in this doc are highly dependent
on the success that we have on-ramping new contributors to the project, and other
items will depend on external contributions, which can vary.

If you are interested in contributing to a particular area, you can look through
the project board for that area and assign yourself to one of the issues. It is
recommended to start with smaller issues to ramp up on the project before starting
to tackle larger issues.

Project boards:
https://github.com/orgs/kubernetes-sigs/projects/50
https://github.com/orgs/kubernetes-sigs/projects/51
https://github.com/orgs/kubernetes-sigs/projects/52
https://github.com/orgs/kubernetes-sigs/projects/53
https://github.com/orgs/kubernetes-sigs/projects/54

## Kustomize contributors (at time of writing):

kustomize owner: @natasha41575

kustomize maintainers: @annasong20, @koba1t

kustomize contributors: @varshaprasad96


# H2 2023

## Goal: Create kustomize leadership and contributor playbooks

Contributors: natasha41575, annasong20

Priority: High

Effort: Medium

In the past, when the leads stopped contributing (for various reasons, not covered here)
in various kubernetes projects, it left a wide hole that few could easily fill,
leaving the remaining leads in a bad position and the project understaffed. We should assume
that we will need to onboard new maintainers in the future, and should have playbooks
for doing so. As we grow the contributor base in kustomize, we will build these playbooks for
those who are contributing and those who are looking to grow into kustomize leaders.
To ensure the long term health and stability of the project, we should have:

- On-boarding guides for new contributors
- Clear guidelines for how to climb the kustomize ladder from contributor to approver to owner
- A plan (maybe a schedule) for future kustomize cohorts
- A succession plan, in case the current kustomize leads ever decide to step down

## Goal: Onboard 2-5 new contributors to kustomize

Contributors: natasha41575, annasong20, koba1t

Priority: High

Effort: High

In order to make progress on kustomize goals in the future, we need to increase the
level of staffing on kustomize. We should leverage community contributions to keep kustomize
healthy and making progress.

The primary means in which we will try to find new kustomize contributors is through the new kustomize
maintainer training cohort. We will lead a group of ~20 kubernetes community members through a 3-6 month
training program, involving talk sessions, bug scrubs, issue triage, PR reviews, and coding projects for
each member.

See [our call for help](https://groups.google.com/a/kubernetes.io/g/dev/c/M5OphEVsv5o/m/zc6G4H15AAAJ) for more
specific details about the program.

The effort from existing kustomize maintainers here will be to:
- Organize the cohort, so that each cohort member feels productive and understands what they should work on
- Align motivation of the cohort members with the work that we assign to them.
- Review PRs from cohort members in a timely manner.
- Be the point(s) of contact for questions/escalations
- Lead weekly stand-ups and monthly bug scrubs

At the end of all this, if we have a small team of contributors to kustomize, who understand its founding
philosophy and intentions, we should be able to keep the project up to date.

## Goal: Improve kustomize extensibility through KRM functions and CRD support

Contributors: koba1t, varshaprasad96, external contributors

Priority: High

Effort: High

Project board: https://github.com/orgs/kubernetes-sigs/projects/53/views/1

For a long time, we have supported KRM functions as the proper way to implement custom generators and transformers.
However, due to limited staffing, we have been unable to drive this feature out of alpha in kustomize. The two
main features which we hope to make progress on are Composition and Catalog, two long-standing proposals for which
numerous users have been waiting for a long time. There are several open issues
regarding KRM functions where our long-term answer has been these two features, but users have been hearing about them
for over a year without seeing any progress. If we can implement them, they will vastly improve usability and security
of KRM functions.

One item that falls under this category that does not currently have a contributor is improving CRD support.
Currently, it is difficult to use CRDs properly, as there are three different fields (configurations, openapi, and crds)
where users have to input their CRD configuration. We need to consolidate these fields into one easy to use feature to
support CRDs. If you are interested in putting together a design proposal for how to tackle this task, please reach
out to the kustomize maintainers.

# H1 2024

## Goal: Improve the kustomize documentation

Contributors: annasong20, external contributions

Priority: High

Effort: High

Project board: https://github.com/orgs/kubernetes-sigs/projects/50

The kustomize documentation is currently fragmented, out of date, and lacks examples to fully understand its value.
We have had a "docs project" for a long time; we need to prioritize implementing it so that the documentation is in
one place, easy to find, and helps new users get started more easily. Some outcomes from this project should be:

- A single, unified website hosted on kustomize.io
- Updated information architecture, and a plan to keep it up to date
- End to end examples of using kustomize, including complex use cases

## Goal: Fix core usability bugs in kustomize

Contributors: external contributions

Priority: High

Effort: High

Project board: https://github.com/orgs/kubernetes-sigs/projects/51

There are several core usability issues that block some users from adopting kustomize features or in
some base block users from using kustomize entirely. These issues range from small bugs with workable but
inconvenient workarounds, to enormous feature gaps.

As part of this goal, we should work toward reducing the number of such issues that we have, making
kustomize work more smoothly and predictably, and be usable for a larger range of users.

There are a lot of important issues in this project, but the biggest and highest priority one is that
kustomize doesn't currently support metadata.GenerateName. Unfortunately, we don't currently have anyone
actively working on this issue, we would need an external contributor to reach out to the kustomize
maintainers to pick it up.

## Goal: Improve kustomize CI, release, & security patch processes

Contributors: external contributions

Priority: Medium

Effort: High

Project board: https://github.com/orgs/kubernetes-sigs/projects/54

The kustomize release process is currently done on-demand and is strictly linear. This means that if we find a CVE,
we are forced to release the next version of kustomize ASAP, and we are required to release every PR that has merged
since the last release. This can put us in a sticky situation if we have a breaking change that we are
not ready to release yet, but we need a patch quickly.

We should try to improve the kustomize release process so that we can release frequently, reliably, and with some
flexibility. The outcome of this effort should be:

- kustomize is released on a regular cadence (biweekly or monthly)
- kustomize is able to separate patch and feature releases, so that we can fix CVEs without needing to release
everything that we have in flight
- We can detect and fix CVEs early

## Goal: Take long-standing alpha commands out of alpha

Contributors: external contributions

Priority: Medium

Effort: Medium

Project board: https://github.com/orgs/kubernetes-sigs/projects/52

There are several commands in kustomize that have been alpha for a long time, including the cfg command group and
localize. Moving them forward can indicate good health of a project and these commands are useful to many users.
Some of these projects can be good starter issues for new contributors to have an easier onramp while others will
require more effort and thought.
