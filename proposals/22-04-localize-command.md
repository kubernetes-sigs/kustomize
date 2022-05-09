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
functions. Downloaded images would live only in local caches and kustomize cannot know the form that custom fields will
take. Thus, neither is worth localizing.

**Goals:**

1. This command should localize 
   * all remote files that a kustomization file directly references 
   * remote exec binaries of referenced KRM functions 

   This command achieves this goal if, in the absence of remote images and custom fields in KRM
   functions, `kustomize build` can run on the localized copy without network access.

**Non-goals:**

1. This command should not localize remote images or custom fields in KRM functions.
2. This command should not copy files that the target kustomization does not reference.
3. This command should not serve as a package manager.

## Proposal

The command takes the following form:

<pre>
<b>kustomize localize</b> <ins>target</ins> <ins>scope</ins> <ins>newDir</ins>
</pre>

where the arguments are:

* `target`: a directory with a top-level kustomization file that kustomize will localize; can be a path to a local
  directory or a url to a remote directory
* `scope`: optional root directory, files outside which kustomize is not allowed to copy and localize; if not specified,
  takes on value of `target`
* `newDir`: destination directory of the localized copy of `target`

The command creates a copy of the `target` kustomization and the local files that `target` references at `newDir`. We
define the "files that `target` references" as:

* kustomization files that `target` directly or transitively references
* configuration files that referenced kustomization files reference
* exec binaries of referenced KRM functions

Here, configuration file means a non-kustomization yaml file. The command only copies referenced files that reside
inside `scope`.

The command localizes the copy of `target` at `newDir` by downloading all remote files that `target` references. The
downloaded files placed in a new `localized-files` directory next to the file that referenced the downloaded files.
Inside `localized-files`, the downloads are located on path:

<pre>
<ins>remote-host</ins> / <ins>organization</ins> / <ins>repo</ins> / <ins>version</ins> / <ins>path-to-file-in-repo</ins>
</pre>

The command rewrites remote references in `newDir` to the local paths of the downloaded files. To help ensure
that `newDir` is a clean copy, the command additionally overwrites absolute path references into `target` to point
to `newDir`.

**Error cases**:
* `target` does not have a top-level kustomization file
* `scope` does not contain `target`
* `newDir` already exists
* `localized-files` directory already exists
* remote reference does not have a version
* kustomization file is malformed
* cycle of kustomization file references exists

**Warning cases**:
* `target` references a local path that traverses outside of `scope`
* KRM function has container image that the user might not have locally

For the warning cases, the command will ignore the reference and continue execution.

If the command runs without any errors or warnings, `kustomize build` without `--load-restrictor LoadRestrictionsNone`
on `target` and `newDir` should produce the same output.

### User Stories

#### Story 1

My company’s CI/CD pipeline currently pulls an `example` directory from our internal package management site. I want the
CI/CD pipeline to additionally run `kustomize build example/overlay`. My setup looks like this:
```
└── example
    ├── overlay
    │   └── kustomization.yaml
    └── base
        └── kustomization.yaml
```
```
# example/overlay/kustomization.yaml
resources:
  - ../base
  - https://raw.githubusercontent.com/kubernetes-sigs/kustomize/v1.0.6/examples/helloWorld/deployment.yaml
```
```
# example/base/kustomization.yaml
resources:
  - github.com/kubernetes-sigs/kustomize/examples/multibases?ref=v1.0.6
  - https://raw.githubusercontent.com/kubernetes-sigs/kustomize/v1.0.6/examples/helloWorld/configMap.yaml
```

I get an error from `kustomize build` in the pipeline because my configurations have remote references, but my CI/CD
pipeline does not have external network access.

Fortunately, I remember that I can run `kustomize localize` on `example/overlay` on my local machine. I can then upload
the localized directory to my company’s internal package management site for the CI/CD pipeline to pull and build
instead. I run
`kustomize localize example/overlay example/localized-overlay`, but I get the following warning:

```
$ kustomize localize example/overlay example/localized-overlay
Warning: File example/overlay/kustomization.yaml refers to ../base on line 2. This reference is outside of scope: 
         example/overlay. kustomize localize will skip this path.
```

because I forgot that `kustomize localize` can only process local references to files within
`scope`. Therefore, the command could not copy `example/base` to `example/localized-overlay`or localize the remote
references in `example/base`. The resulting file structure is as follows:

```
└── example
    ├── overlay
    │   └── kustomization.yaml
    ├── localized-overlay
    │   ├── kustomization.yaml
    │   └── localized-files
    │       └── github.com
    │           └── kubernetes-sigs
    │               └── kustomize
    │                   └── v1.0.6
    │                       └── examples
    │                           └── helloWorld
    │                               └── deployment.yaml
    └── base
        └── kustomization.yaml
```

`kustomize build example/localized-overlay` will still run correctly outside the CI/CD pipeline because I chose to
place `example/localized-overlay` in the same directory as `example/overlay`. As a result, relative paths,
namely `../base`, in `example/localized-overlay` will point to the same files as their counterparts in `example/overlay`
. However, `kustomize build example/localized-overlay` will still require network access to run.

#### Story 2

I am back again from **Story 1**, but this time ready to localize `base` too, instead of just the `overlay` directory.
To localize both, I set my `scope` argument to `example` to make a copy of both directories. My setup still looks like
that at the end of **Story 1**:

```
└── example
    ├── overlay
    │   └── kustomization.yaml
    ├── localized-overlay
    │   ├── kustomization.yaml
    │   └── localized-files
    │       └── github.com
    │           └── kubernetes-sigs
    │               └── kustomize
    │                   └── v1.0.6
    │                       └── examples
    │                           └── helloWorld
    │                               └── deployment.yaml
    └── base
        └── kustomization.yaml
```

After I run `kustomize localize example/overlay example new-space/localized-example`, I get the
following:

```
├── example                         # old kustomization directory
│   ├── overlay
│   │   └── kustomization.yaml
│   ├── localized-overlay
│   │   ├── kustomization.yaml
│   │   └── localized-files
│   │       └── github.com
│   │           └── kubernetes-sigs
│   │               └── kustomize
│   │                   └── v1.0.6
│   │                       └── examples
│   │                           └── helloWorld
│   │                               └── deployment.yaml
│   └── base
│       └── kustomization.yaml
└── new-space
    └── localized-example            # the new, localized kustomization directory
        ├── kustomization.yaml
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
        │                               ├── kustomization.yaml
        │                               ├── production
        │                               └── staging
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
```
# new-space/localized-example/overlay/kustomization.yaml
resources:
  - ../base
  - ./localized-files/github.com/kubernetes-sigs/kustomize/examples/helloWorld/deployment.yaml
```
```
# new-space/localized-example/base/kustomization.yaml
resources:
  - ./localized-files/github.com/kubernetes-sigs/kustomize/examples/multibases
  - ./localized-files/github.com/kubernetes-sigs/kustomize/examples/helloWorld/configMap.yaml
```

Now, I upload `new-space/localized-example` from my local setup to my company’s internal package management site. I
change the commands in my CI/CD pipeline to pull `localized-example` before running `kustomize build localized-example`,
and the command executes successfully!

### Risks and Mitigations
N/A

### Dependencies
N/A

### Scalability

A chain of remote kustomization directories in which the current kustomization file references the next remote
kustomization directories could create a `newDir` with deeply nested `localized-files`. This directory structure would
impede users’ navigation of `newDir`. However, this scenario should be unlikely as most kustomizations only consist of a
few layers.

The creation of the `localized-files` directory local to the referencing kustomization file additionally prevents the
different layers of kustomization files from sharing the same copy of the remote files. Following the same logic,
different potential `target` directories cannot share copies either.

## Drawbacks

Users whose layered kustomizations form a complex directory tree structure may have a hard time finding an
appropriate `scope`. However, many kustomizations exist in repositories, allowing the user to easily choose the repo
root as a valid `scope`. The warning messages that `kustomize localize` outputs for reference paths that extend
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
  complicates the existing kustomize workflow by requiring the setup of global environment variables.
  <br></br>

* The command could, instead of making a copy, modify `target` directly. However, users would not have an easy way to
  undo the command, which is undesirable.
  <br></br>

* Instead of requiring the user to specify a second argument `scope`, the command could by definition limit its copying
  to `target`. However, in the case of **Story 1**, the command would force the user to set `target` to `example` in
  order to include `example/base` in the localization of `example/overlay`. The user would then have to create a
  kustomization file at `example` that points to `example/overlay` under the `resources` field. The creation of the
  kustomization file solely for this purpose is messy and more work for the user.

## Rollout Plan

This command will have at least alpha and GA releases. Depending on user feedback, we may add a beta.

### Alpha

This release will ignore 

* local kustomization files that `target` references
* KRM functions

The entire command will be new in the alpha release, and so will not require an alpha flag. The command will not be
available in `kubectl kustomize` either as kubectl only has `kustomize build` builtin.

### Beta/GA

This release should have all features documented in this proposal. Though, we may make changes based on user feedback.
