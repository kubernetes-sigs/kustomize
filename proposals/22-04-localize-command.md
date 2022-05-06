# Localize Command

**Authors**:

- annasong20

**Reviewers**:

- natasha41575
- KnVerey

**Status**: implementable

## Summary

The `kustomize localize` command creates a “localized” copy of the target kustomization and any files target transitively
references, in which the kustomization files contain local, instead of remote, references to downloaded files. The
command is part of an effort to enable `kustomize build` to run without network access.

## Motivation

monopole originally proposed the command `kustomize localize`
in [this issue](https://github.com/kubernetes-sigs/kustomize/issues/3980).

Users run `kustomize build` in many environments with limited network access. For example, CI/CD pipelines often only
have access to the internal network. Server-side applications like Config Sync are concerned with the security
vulnerabilities of git, which `kustomize build` uses to fetch remote files.

These use cases would benefit from a kustomize solution that downloads all remote files that a `kustomize build` target
references, into a copy of target that references the downloaded files instead. The copy would include the target and
any files target transitively references.`kustomize build` would then be able to run on the copy without a network
dependency.

This proposal nearly achieves the solution by downloading all remote files directly referenced by the target or by a
transitively referenced kustomization file. The only remote files not covered by this proposal and still needed for
`kustomize build` to run are those in KRM functions. The command copies local exec binaries of KRM functions only as
part of copying target. The actual localization only applies to kustomization resources. KRM functions are third party
to kustomize, and thus KRM function remotes are out of scope for `kustomize localize`.

**Goals:**

1. This command should localize all remote files that a kustomization file directly references. This command will have
   achieved this goal if, in the absence of remote input files to KRM functions, `kustomize build` can run on the
   localized copy without network access.

**Non-goals:**

1. This command should not localize remote input files to KRM functions.
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
* `scope`: `target` or a directory that contains `target`
* `newDir`: destination of the localized copy of `target`
    * if `target` is local, `newDir` must be a directory name, as it will be located in the same directory as `target`
      to preserve relative path references
    * if `target` is remote, `newDir` must be a directory path

The command creates a copy of `target` at `newDir` in which each kustomization file, from the top-level to any
recursively referenced, has local paths to downloaded files instead of remote references for the following kustomization
fields:

* `resources`
* `components`
* `bases`
* `openapi:`  
  &nbsp; &nbsp; `paths`

A new `localized-files` directory holds the downloaded files (or directory)
at:

<pre>
<ins>remote-host</ins> / <ins>organization</ins> / <ins>repo</ins> / <ins>version</ins> / <ins>path-to-file-in-repo</ins>
</pre>

Each `localized-files` directory is located in the same directory as the kustomization file that referenced the
downloaded files.

To help ensure that `newDir` is a clean copy, the command overwrites every absolute path into `target` to point
to `newDir` before processing the path.

**Error cases**:
* `target` does not have a top-level kustomization file
* `newDir` already exists
* `localized-files` directory already exists
* remote reference does not have a version
* kustomization file is malformed
* cycle of kustomization file references exists

**Warning cases**:
* `newDir` refers to `base` kustomization layer outside of `newDir`, and `base` has remote references
* KRM function has container image or exec binary that the user might not have locally

If the command runs without any errors, `kustomize build` on `target` and `newDir` should produce the same output.

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
`kustomize localize example/overlay example/localized-overlay`, but I get the following error:

```
$ kustomize localize example/overlay example/localized-overlay
Warning: kustomization directory example/overlay refers to remote resources in example/base
```

because I forgot that `kustomize localize` can only localize remote references originating from within the
`target`, `example/overlay`. Therefore, command could not localize the remote references in `example/base`,
which `example/overlay/kustomization.yaml` locally references. `kustomize build example/localized-overlay` will still
run correctly outside the CI/CD pipeline, but will still require network access:

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

#### Story 2

I am back again from **Story 1**, but this time ready to localize `base` too, instead of just the `overlay` directory.
To localize both, I change my `target` argument from `example/overlay` to `example` to make a copy of both directories.
I add a top-level kustomization file to `example` with the following commands:

```
kustomize init; kustomize edit add resource overlay
```

I have deleted `example/localized-overlay` from **Story 1**. My setup now looks like this:
```
└── example
    ├── kustomization.yaml
    ├── overlay
    │   └── kustomization.yaml
    └── base
        └── kustomization.yaml
```
```
# example/kustomization.yaml
resources:
  - ./overlay
```

with all other files having the same contents. After I run `kustomize localize example localized-example`, I get the
following:
```
├── example                      # the old kustomization directory
│   ├── kustomization.yaml
│   ├── overlay
│   │   └── kustomization.yaml
│   └── base
│       └── kustomization.yaml
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
# localized-example/overlay/kustomization.yaml
resources:
  - ../base
  - ./localized-files/github.com/kubernetes-sigs/kustomize/examples/helloWorld/deployment.yaml
```
```
# localized-example/base/kustomization.yaml
resources:
  - ./localized-files/github.com/kubernetes-sigs/kustomize/examples/multibases
  - ./localized-files/github.com/kubernetes-sigs/kustomize/examples/helloWorld/configMap.yaml
```

Now, I upload `localized-example` from my local setup to my company’s internal package management site. I change the
commands in my CI/CD pipeline to pull `localized-example` before running `kustomize build localized-example`, and the
command executes successfully!

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

Users, like the one in the **User Stories** section, whose kustomization layers are in sibling directories need to
perform the extra step of creating a top-level kustomization file. However, many existing kustomize use cases also
require this step, and as shown in **Story 2**, users can create kustomization files relatively easily via command line.

## Alternatives

* Instead of downloading into `newDir`, `kustomize localize` could download all remote files into a directory specified
  by some global environment variable (like in Golang), which would preclude deeply nested directories and allow
  different kustomization files to share configurations. On top of that, if `kustomize build` had the added
  functionality to check for previous downloads of remote references at said global location, `kustomize localize` would
  not need to overwrite the remote references in `target` to the local downloads. As a result, `kustomize localize`
  would need to neither write to `target` nor copy `target` into `newDir`. The user would not need to create a top-level
  kustomization file either. <br></br>

  Despite its advantages, the alternative design violates the self-contained nature of each kustomize layer. Users would
  be unable to upload a fully localized kustomization directory in version control. Furthermore, this alternative
  complicates the existing kustomize workflow by requiring the setup of global environment variables.

* The command could, instead of making a copy, modify `target` directly. However, users would not have an easy way to
  undo the command, which is undesirable.

## Rollout Plan
This command will have at least alpha and GA releases. Depending on user feedback, we may add a beta.

### Alpha

This release will limit the depth of nested `localized-files` to 1 layer. In other words, the command will ignore remote
kustomization directory references that originate from within a `localized-files` directory. In addition, this release
will ignore KRM functions. The command will not output warnings for remote KRM images or exec binaries.

The entire command will be new in the alpha release, and so will not require an alpha flag. The command will not be
available in `kubectl kustomize` either as kubectl only has `kustomize build` builtin.

### Beta/GA

This release should have all features documented in this proposal. Though, we may make changes based on user feedback.
