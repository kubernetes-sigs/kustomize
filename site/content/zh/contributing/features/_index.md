---
title: "Contributing Features"
linkTitle: "Contributing Features"
type: docs
weight: 21
description: >
    How to contribute features
---

[issue]: https://github.com/kubernetes-sigs/kustomize/issues
[sig-cli]: /kustomize/contributing/community/
[meeting agenda]: https://docs.google.com/document/d/1r0YElcXt6G5mOWxwZiXgGu_X6he3F--wKwg-9UBc29I/edit#heading=h.himo1st0tqyy
[KEP]: https://github.com/kubernetes/enhancements/tree/master/keps/sig-cli
[table-driven]: https://github.com/kubernetes-sigs/kustomize/blob/a8b9741866cf8e0c43e643ab7a9f40a3bd7e2a4d/api/filters/imagetag/imagetag_test.go#L15
[eschewed feature list]: https://kubernetes-sigs.github.io/kustomize/faq/eschewedfeatures/
[kind/feature]: https://github.com/kubernetes-sigs/kustomize/labels/kind%2Ffeature

Following is the process for proposing a new Kustomize feature:

1. Check the [eschewed feature list] to see if the feature has already been proposed
2. File an [issue] describing the desired feature
   - label it [kind/feature]
   - the motivation for the feature
   - example of how you would accomplish the motivating task *without* the feature
   - example of how you would accomplish the motivating task *with* the feature
3. Email the [sig-cli] mailing list with the issue
4. Present the issue at [sig-cli] bi-weekly meeting on Zoom
   - add it to the [meeting agenda] doc
   - be present to discuss the feature
   - response may be -- move forward with a PoC, not to move forward, defer and come back later,
     or more information is needed.
5. Address the feedback on the issue
   - Possibly write a KEP for tracking the feature
6. Implement the feature and send a PR
   - Add [table-driven] tests
   - Expect comments on the PR within 2 weeks
7. Kustomize team will release the kustomize `api` and `kustomize` modules
