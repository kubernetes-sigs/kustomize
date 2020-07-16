---
title: "Filing Bugs"
linkTitle: "Filing Bugs"
type: docs
weight: 10
description: >
    How to file bugs and fix Kustomize bugs
---


[krusty package]: https://github.com/kubernetes-sigs/kustomize/tree/master/api/krusty
[reusable custom transformer test]: https://github.com/kubernetes-sigs/kustomize/tree/master/api/krusty/customconfigreusable_test.go

File issues as desired, but if you've found a problem
with how `kustomize build` works, please report

* the output of `kustomize version`,
* the input (the content of `kustomization.yaml`
   and any files it refers to),
* the expected YAML output.

## If you have `go` installed

kustomize has a simple test harness in the [krusty
package] for specifying a kustomization's input and the
expected output.

Copy one of those tests, e.g. this [reusable custom
transformer test], to a new test file in the
krusty package.

Insert the inputs you want to use, and run it as
you'd run the reusable custom transformer test:

```
(cd api; go test -run TestReusableCustomTransformers ./krusty)
```

The output will demonstrate the bug or missing feature.

Record this output in the test file in a call to
`AssertActualEqualsExpected`, per all the other tests
in the [krusty package].  This makes the test pass,
albeit with output demonstrating behavior you
presumably want to change.

Send the new test in a PR, along with commentary (in
the test) on what you'd prefer to see.

The person who fixes the bug then has a clear bug
reproduction and a test to modify when the bug is
fixed.

Any bug fix first requires a test demonstrating the bug
(so we have permanent regression coverage), so if the
_bug reporter_ does this, it saves time and avoids
misunderstandings.
