# Filing bugs

[target package]: https://github.com/kubernetes-sigs/kustomize/tree/master/pkg/target
[example of a target test]: https://github.com/kubernetes-sigs/kustomize/blob/master/pkg/target/baseandoverlaysmall_test.go

File issues as desired, but
if you've found a problem with how
`kustomize build` works, consider the
following to improve response time.

## A good report specifies

 * the output of `kustomize version`,
 * the input (the content of `kustomization.yaml`
   and any files it refers to),
 * the expected YAML output.

## A great report is a bug reproduction test

kustomize has a simple test harness in the
[target package] for specifying a kustomization's
input and the expected output.

See this [example of a target test], and contribution
[#971](https://github.com/kubernetes-sigs/kustomize/pull/971),
which does exactly the right thing.

The pattern is
 * call `NewKustTestHarness`
 * specify kustomization input data (resources,
   patches, etc.) as inline strings,
 * call `makeKustTarget().MakeCustomizedResMap()`
 * compare the actual output to expected output

In a bug reproduction test, the expected output
string initially contains the _wrong_ (unexpected)
output, thus unambiguously reproducing the bug.

Nearby comments should explain what the output
should be, and have a TODO pointing to the related
issue.

The person who fixes the bug then has a clear bug
reproduction and a test to modify when the bug is
fixed.

The bug reporter can then see the bug was fixed,
and has permanent regression coverage to prevent
its reintroduction.
