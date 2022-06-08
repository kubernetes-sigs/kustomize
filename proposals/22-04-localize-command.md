# Localize Command

**Authors**:

- annasong20

**Reviewers**:

- natasha41575
- KnVerey

**Status**: implementable

## Summary

The `kustomize localize` command creates a “localized” copy, of both the target kustomization and files target
references, in which the kustomization files contain, instead of remote references, local paths to their downloaded
locations. The command is part of an effort to enable `kustomize build` to run without network access.

## Motivation

monopole originally proposed the command `kustomize localize`
in [this issue](https://github.com/kubernetes-sigs/kustomize/issues/3980).

Users run `kustomize build` in many environments with limited network access. For example, CI/CD pipelines often only
have access to the internal network. Server-side applications like Config Sync are concerned with the security
vulnerabilities of git, which `kustomize build` uses to fetch remote files.

These use cases would benefit from a kustomize solution that downloads all remote files that a `kustomize build` target
references, into a copy of target that references the downloaded files instead. Admins could upload the localized copy
to an internal repo so that pipelines and applications can run `kustomize build` on the copy without a network
dependency.

The proposed command nearly achieves the solution by downloading all remote files directly referenced by the target or
by a recursively referenced kustomization file. The command also downloads remote exec binaries of referenced KRM
functions, which are the only potential source of remote files other than kustomizations. The only remote files that
this proposal does not cover and that `kustomize build` still needs to run are remote images and custom fields in KRM
functions. Downloaded images would live only in local caches and thus, are not worth localizing. Kustomize cannot
currently identify custom fields, though this may change with one of the proposed solutions
in [issue #4154](https://github.com/kubernetes-sigs/kustomize/issues/4154).

The proposed command has the added benefit of increasing user confidence in the integrity of their kustomization builds.
Locally downloaded files, unlike urls, give users full control of file content. At the same time, the command does this
without modifying the original kustomization so that users can always run the command on the original again to fetch
upstream changes.

**Goals:**

This command should localize

* all remote files that a kustomization file directly references
* remote exec binaries of referenced KRM functions

This command achieves this goal if, in the absence of remote images and custom fields in KRM
functions, `kustomize build` can run on the localized copy without network access and produce the same output as when
run on the original.

**Non-goals:**

1. This command should not localize remote images or custom fields in KRM functions.
2. This command should not copy files that the target kustomization does not reference.
3. This command should not serve as a package manager.

## Proposal

The command takes the following form:

<pre>
<b>kustomize localize</b> <ins>target</ins> <ins>newDir</ins> [--scope <ins>scope</ins>] [--no-verify]
</pre>

where the arguments are:

* `target`: a directory with a top-level kustomization file that kustomize will localize; can be a path to a local
  directory or a url to a remote directory
* `newDir`: optional destination directory of the localized copy of `target`; if not specified, the destination is a
  directory named `localized-{target}` in the working directory

and the flags are:
* `--scope scope`: optional root directory, files outside which kustomize is not allowed to copy and localize; if not
  specified, takes on value of `target`
* `--no-verify`: do not verify that the outputs of `kustomize build` for `target` and `newDir` are the same after
  localization

The command creates a copy of the `target` kustomization and the local files that `target` references at `newDir`. We
define the "files that `target` references" as:

* kustomization files that `target` directly or transitively references
* configuration files that referenced kustomization files reference
* exec binaries of referenced KRM functions

Here, configuration file means a non-kustomization yaml file. The command cannot run on `target`s that need
the `--load-restrictor LoadRestrictionsNone` flag for `kustomize build`. Additionally, the command only copies
referenced files that reside inside `scope`. Note that for the localization to occur, `scope` must contain `target` and
naturally if `target` is remote, `scope` should be as well. The copied files sit under the same relative paths
in `newDir` as those that their counterparts sit under in `scope`.

The command localizes the copy of `target` in `newDir` by downloading all remote files that `target` references. For
each downloaded exec binary that KRM functions reference, the command removes users' executable permissions and prints a
warning message to that effect.

The command creates a new `localized-files` directory, next to the file that referenced the downloaded files, to hold
said files. Inside `localized-files`, the downloads are located on path:

<pre>
<ins>domain</ins> / <ins>organization</ins> / <ins>repo</ins> / <ins>version</ins> / <ins>path/to/file/in/repo</ins>
</pre>

where `version` corresponds to a [`git fetch` "ref"](https://git-scm.com/docs/git-fetch), found in the form of the `ref`
query string parameter for directory urls and embedded in the path of raw GitHub file urls. Ideally though, `version` is
a stable tag as opposed to a branch.

The command replaces remote references in `newDir` with local relative paths to the downloaded files. To help ensure
that `newDir` is a clean copy, the command additionally overwrites absolute path references into `target` to point
to the corresponding file in `newDir` instead.

As a convenience to the user, in the absence of the `--no-verify` flag, the command automatically tries to
run `kustomize build`, without any flags, on the original `target` and the localized `target` in `newDir` to compare
their outputs. The command indicates success if the outputs match and throws an error with a diff summary otherwise.
This check, however, is not useful for certain `target`s, including those that need flags to build. In these cases, the
command prints next steps that users can follow to check the output themselves. For example, for `target`s that
reference KRM functions with a remote exec binary, the command suggests the user:

1. add executable permissions for the downloaded exec binaries in `newDir` **that the user trusts**
2. run `kustomize build` with flags `--enable-alpha-plugins --enable-exec` and self-verify the outputs

**Error cases**:

* `target` does not have a top-level kustomization file
* `kustomize build` needs `--load-restrictor LoadRestrictionsNone` to run on `target`
* `newDir` already exists
* `scope` does not contain `target`
* `target` references a local path that traverses outside of `scope`
* remote reference does not have a `version`
* `localized-files` directory already exists
* kustomization file is malformed
* cycle of kustomization file references exists
* `kustomize build` produces different output for `target` and `newDir` in the absence of `--no-verify`

Depending on feedback, we may add an `--overwrite` flag in the future to allow users to update an existing `newDir` by
running the command again.

**Warning cases**:

* KRM function references remote exec binary, in which case the downloaded exec binary is not executable
* KRM function has container image that the user might not have locally

### User Stories

#### Story 1

My company’s CI/CD pipeline currently pulls an `example` directory from our internal package management site. I want the
CI/CD pipeline to additionally run `kustomize build example/overlay`. My setup looks like this:
```shell
└── example
    ├── overlay
    │   └── kustomization.yaml
    └── base
        └── kustomization.yaml
```
```shell
# example/overlay/kustomization.yaml
resources:
  - ../base
  - https://raw.githubusercontent.com/kubernetes-sigs/kustomize/v1.0.6/examples/helloWorld/deployment.yaml
```
```shell
# example/base/kustomization.yaml
resources:
  - https://github.com/kubernetes-sigs/kustomize//examples/multibases?ref=v1.0.6
  - https://raw.githubusercontent.com/kubernetes-sigs/kustomize/v1.0.6/examples/helloWorld/configMap.yaml
```

I get an error from `kustomize build` in the pipeline because my configurations have remote references, but my CI/CD
pipeline does not have external network access.

Fortunately, I remember that I can run `kustomize localize` on `example/overlay` on my local machine. I can then upload
the localized directory to my company’s internal package management site for the CI/CD pipeline to pull and build
instead. I run `kustomize localize example/overlay --scope example`, where my `target` is `example/overlay`, I accept
the default location and name of `newDir`, and I expand my `scope` to `example` because `example/overlay`
references `example/base`. I get the following output:

```shell
$ kustomize localize example/overlay --scope example
SUCCESS: example/overlay, localized-overlay produce same kustomize build output
```

```shell
├── example                         # old kustomization directory
│   ├── overlay
│   │   └── kustomization.yaml
│   └── base
│       └── kustomization.yaml
└── localized-overlay               # the new, localized kustomization directory
    ├── base
    │   ├── kustomization.yaml
    │   └── localized-files
    │       └── github.com
    │           └── kubernetes-sigs
    │               └── kustomize
    │                   └── v1.0.6
    │                       └── examples
    │                           ├── helloWorld
    │                           │   └── configMap.yaml
    │                           └── multibases
    │                               ├── base
    │                               │   └── pod.yaml
    │                               ├── dev
    │                               │   └── kustomization.yaml
    │                               ├── kustomization.yaml
    │                               ├── production
    │                               │   └── kustomization.yaml
    │                               └── staging
    │                                   └── kustomization.yaml
    └── overlay
        ├── kustomization.yaml
        └── localized-files
            └── github.com
                └── kubernetes-sigs
                    └── kustomize
                        └── v1.0.6
                            └── examples
                                └── helloWorld
                                    └── deployment.yaml
```
```shell
# localized-overlay/overlay/kustomization.yaml
resources:
  - ../base
  - localized-files/github.com/kubernetes-sigs/kustomize/v1.0.6/examples/helloWorld/deployment.yaml
```
```shell
# localized-overlay/base/kustomization.yaml
resources:
  - localized-files/github.com/kubernetes-sigs/kustomize/v1.0.6/examples/multibases
  - localized-files/github.com/kubernetes-sigs/kustomize/v1.0.6/examples/helloWorld/configMap.yaml
```

Now, I upload `localized-overlay` from my local setup to my company’s internal package management site. I change the
commands in my CI/CD pipeline to pull `localized-overlay` before running `kustomize build localized-overlay`, and the
command executes successfully!

### Risks and Mitigations

One could argue that while uploading the localized `newDir` to a repository, a user could accidentally leak Secrets that
were originally remote in a more private repo. This event is not very likely, as the servers that the user intends to
consume the localized kustomizations often only have access to internal, private networks and repos. Nonetheless, users
should have enough context in the case of Secrets to make the right decision. Users should have a basic understanding of
the files that their `target` kustomizations reference and of the files that they plan to upload to repos. Secret
configurations are also not too difficult to identify.

Exec binaries that KRM functions reference are a different story. `kustomize localize` downloads remote exec binaries
that, if malicious, are capable of almost anything during subsequent `kustomize build` calls. The command mitigates this
risk by removing executable permissions on these downloaded exec binaries and warning the user, as mentioned in 
**Proposal**. `kustomize build` can only run the exec binary after the user deems the binary safe and changes its
permissions.

Still another risk may be that if a user's kustomization tree is large, `kustomize localize` has the potential to copy
files from unexpected locations. The command mitigates this risk with the `scope` flag. If the user specifies `scope`,
they understand that `kustomize localize` only copies files in `scope`. Otherwise, `kustomize localize` treats `target`
as `scope`. In either case, `kustomize localize` aborts with a descriptive error message if `target` references local
files outside of `scope` that the command would copy. The qualification that the command can only localize `target`s
that follow load restrictions helps mitigate this risk as well.

### Dependencies

N/A

### Scalability

Large kustomization trees slow the performance of `kustomize localize`. These trees can have large local subtrees, have
large remote subtrees, be deeply nested, or be wide, with each overlay referencing multiple bases. Regardless of the
cause, large kustomization trees inevitably take longer to copy and download. Parts of the kustomize code are not
thread-safe, which precludes parallel execution.

The creation of the `localized-files` directory local to the referencing kustomization file additionally prevents the
different layers of kustomization files from sharing the same copy of the remote files. Following the same logic,
different potential `target` directories cannot share copies either.

## Drawbacks

Users whose layered kustomizations form a complex directory tree structure may have a hard time finding an
appropriate `scope`. However, many kustomizations exist in repositories, allowing the user to easily choose the repo
root as a valid `scope`. The error messages that `kustomize localize` outputs for reference paths that extend
beyond `scope` should also help.

## Alternatives

* Instead of downloading into `newDir`, `kustomize localize` could download all remote files into a directory specified
  by some global environment variable (like in Golang), which would preclude deeply nested directories and allow
  different kustomization files to share configurations. On top of that, if `kustomize build` had the added
  functionality to check for previous downloads of remote references at said global location, `kustomize localize` would
  not need to overwrite the remote references in `target` to the local downloads. As a result, `kustomize localize`
  would need to neither write to `target` nor copy `target` into `newDir`. The user would not need to specify a `scope`
  either. <br></br>

  Despite its advantages, the alternative design violates the self-contained nature of each kustomize layer. Users would
  be unable to upload a fully localized kustomization directory in version control. Furthermore, this alternative
  complicates the existing kustomize workflow by requiring the setup of global environment variables. <br></br>

* The command could, instead of making a copy, modify `target` directly. However, users would not have an easy way to
  undo the command, which is undesirable. <br></br>

* Instead of requiring the user to specify a second argument `scope`, the command could by definition limit its copying
  to `target`. However, in the case of **Story 1**, the command would force the user to set `target` to `example` in
  order to include `example/base` in the localization of `example/overlay`. The user would then have to create a
  kustomization file at `example` that points to `example/overlay` under the `resources` field. The creation of the
  kustomization file solely for this purpose is messy and more work for the user.

## Rollout Plan

This command will have at least alpha and GA releases. Depending on user feedback, we may add a beta.

### Alpha

This release will not support

* KRM functions
* absolute paths
* any verification in the form of the `--no-verify` flag or automatically running `kustomize build` at the end of
  localization

The entire command will be new in the alpha release, and so will not require an alpha flag. The command will not be
available in `kubectl kustomize` either, as kubectl only has `kustomize build` builtin. Instead, upon execution, the
command will print a warning message, which declares its alpha status and includes instructions on how users can provide
feedback.

### Beta/GA

This release should have all features documented in this proposal. Though, we may make changes based on user feedback.
