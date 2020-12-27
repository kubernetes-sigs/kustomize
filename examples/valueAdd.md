# Simple addition of one string value

[DRY]: https://en.wikipedia.org/wiki/Don%27t_repeat_yourself

Suppose you have several distinct cloud _projects_
(on GCP or AWS or whatever) named:

* cat-111
* dog-222
* fox-333

These might be project names within the company,
cloud billing identifiers, or both.

Further suppose

* You want to deploy these projects to different
  k8s namespaces, named after the projects.

* You need to specify the project name
  in various resource subfields.

* You want to name the configuration
  directories using the project name.

Additionally you might want to deploy the
projects one at a time, or all at once.

Ideally, you'll want to avoid specifying
the project name in more than one place
(i.e. you want to stay [DRY]).

Here's a possible layout:


> ```
> ├── all
> │   └── kustomization.yaml
> │
> ├── bases
> │   └── iam-iap-tunnel
> │       ├── kustomization.yaml
> │       └── policymembers.yaml
> │
> └── projects
>     ├── cat-111
>     │   └── kustomization.yaml
>     ├── dog-222
>     │   └── kustomization.yaml
>     └── fox-333
>         └── kustomization.yaml
> ```

This layout allows each project to be
individually buildable:

> ```
> kustomize build projects/cat-111
> kustomize build projects/dog-222
> kustomize build projects/fox-333
> ```

or collectively buildable:

> ```
> kustomize build all
> ```

-----
Make a place to work:

<!-- @makePlaceToWork @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)
```

<!-- @defineLayout @testAgainstLatestRelease -->
```
mkdir -p $DEMO_HOME/bases/iam-iap-tunnel
mkdir -p $DEMO_HOME/transformers/setProject
mkdir -p $DEMO_HOME/projects/cat-111
mkdir -p $DEMO_HOME/projects/dog-222
mkdir -p $DEMO_HOME/projects/fox-333
```

To ground this example with a common problem,
assume a set of engineers:

* red@example.com
* blue@example.com
* yellow@example.com

who need particular access to one or more projects.

Define an instance of `IAMCustomRole`:

<!-- @customRole @testAgainstLatestRelease -->
```
cat <<'EOF' >$DEMO_HOME/bases/iam-iap-tunnel/customroles.yaml
apiVersion: iam.cnrm.cloud.google.com/v1beta1
kind: IAMCustomRole
metadata:
  name: engineer
spec:
  title: Colorful Engineer
  permissions:
  - iap.tunnelInstances.accessViaIAP
  stage: GA
EOF
```

Define corresponding instances of `IAMPolicyMember`.

The values in the `resourceRef/external` fields in these instances
are only partially complete.  kustomize will add projectIds to
these below.

The boilerplate in these instances could be eliminated
by making a _custom generator_, but that's for
different tutorial, and with only three instances here
it's not worth it the effort.

<!-- @policyMembers @testAgainstLatestRelease -->
```
cat <<'EOF' >$DEMO_HOME/bases/iam-iap-tunnel/policymembers.yaml
apiVersion: iam.cnrm.cloud.google.com/v1beta1
kind: IAMPolicyMember
metadata:
  name: iap-tunnel-red
spec:
  member: user:red@example.com
  role: roles/engineer
  resourceRef:
    apiVersion: resourcemanager.cnrm.cloud.google.com/v1beta1
    kind: Project
    external: projects
---
apiVersion: iam.cnrm.cloud.google.com/v1beta1
kind: IAMPolicyMember
metadata:
  name: iap-tunnel-blue
spec:
  member: user:blue@example.com
  role: roles/engineer
  resourceRef:
    apiVersion: resourcemanager.cnrm.cloud.google.com/v1beta1
    kind: Project
    external: projects
---
apiVersion: iam.cnrm.cloud.google.com/v1beta1
kind: IAMPolicyMember
metadata:
  name: iap-tunnel-yellow
spec:
  member: user:yellow@example.com
  role: roles/engineer
  resourceRef:
    apiVersion: resourcemanager.cnrm.cloud.google.com/v1beta1
    kind: Project
    external: projects
EOF
```

Make a base that combines these:

<!-- @makeBase @testAgainstLatestRelease -->
```
cat <<'EOF' >$DEMO_HOME/bases/iam-iap-tunnel/kustomization.yaml
resources:
- customroles.yaml
- policymembers.yaml
EOF
```

Make a transformer configuration file.

The transformer used is called `AddValueTransformer`.  It's
intended to implement the 'add' operation of
[IETF RFC 6902 JSON].   The add operation is simple declaration
of what value to add, and a powerful syntax for specifying where
to add the value.  The value can, for example, be inserted
into an existing field holding a file path as either a prefix,
a postfix, or some change
in the middle (e.g. `/volume/data` becomes `/volume/projectId/data`).

[IETF RFC 6902 JSON]: https://tools.ietf.org/html/rfc6902


At the time of writing, this transformer has no dedicated keyword
in the kustomization file to hold it's config.  This means
the config must live in its own file:

<!-- @defineSetProjectTransformer @testAgainstLatestRelease -->
```
cat <<'EOF' >$DEMO_HOME/transformers/setProject/setProject.yaml
apiVersion: builtin
kind: ValueAddTransformer
metadata:
  name: dirNameAdd

#  Omitting the 'value:' field means that the current
#  kustomization root directory name will be used as
#  the value.
# value:  not specified!

targets:
- fieldPath: metadata/namespace
- selector:
    kind: IAMPolicyMember
  fieldPath: spec/resourceRef/external
  filePathPosition: 2
EOF
```

This file defined both the value to insert, and a list of places to
insert it.  It's saying 1) _take the name of the directory I am in_ and
2) use the name as a namespace on all objects in scope, and 3) add that
name to the 2nd position in the file path found in the `spec/resourceRef/external`
field of all `IAMPolicyMember` instances.


To be used, this transformer config file must be referenced
from some kustomization file's `transformers:` field.

This field can contain a path directly to the transformer config file,
or a path to an encapsulating kustomization root.  The latter approach
allows any number of transformers to be loaded as a group from a local
or remote location.

Here an example of the latter case that uses a kustomization file to
list pointers to transformer configs, although in this case it
references only one transformer config.

<!-- @makeTransformerDir @testAgainstLatestRelease -->
```
cat <<'EOF' >$DEMO_HOME/transformers/setProject/kustomization.yaml
resources:
- setProject.yaml
EOF
```

Now make the _cat_, _dog_ and _fox_ _variants_.

These are the targets that one could
independently apply to a cluster.

<!-- @defineCat @testAgainstLatestRelease -->
```
cat <<'EOF' >$DEMO_HOME/projects/cat-111/kustomization.yaml
resources:
- ../../bases/iam-iap-tunnel
transformers:
- ../../transformers/setProject
EOF
```

<!-- @defineDog @testAgainstLatestRelease -->
```
cat <<'EOF' >$DEMO_HOME/projects/dog-222/kustomization.yaml
resources:
- ../../bases/iam-iap-tunnel
transformers:
- ../../transformers/setProject
EOF
```

<!-- @defineFox @testAgainstLatestRelease -->
```
cat <<'EOF' >$DEMO_HOME/projects/fox-333/kustomization.yaml
resources:
- ../../bases/iam-iap-tunnel
transformers:
- ../../transformers/setProject
EOF
```

Then, optionally, a target to deploy all the projects at once:

<!-- @defineAllTarget @testAgainstLatestRelease -->
```
mkdir -p $DEMO_HOME/all
cat <<'EOF' >$DEMO_HOME/all/kustomization.yaml
resources:
- ../projects/cat-111
- ../projects/dog-222
- ../projects/fox-333
EOF
```

The layout is now:

<!-- @showLayout -->
```
tree $DEMO_HOME
```

It should look like:

> ```
> /tmp/someTmpDir
> ├── all
> │   └── kustomization.yaml
> ├── bases
> │   └── iam-iap-tunnel
> │       ├── customroles.yaml
> │       ├── kustomization.yaml
> │       └── policymembers.yaml
> ├── projects
> │   ├── cat-111
> │   │   └── kustomization.yaml
> │   ├── dog-222
> │   │   └── kustomization.yaml
> │   └── fox-333
> │       └── kustomization.yaml
> └── transformers
>     └── setProject
>         ├── kustomization.yaml
>         └── setProject.yaml
> ```

The expected output from building the dog project
is as follows:

<!-- @definedExpectedOutput @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/out_expected.yaml
apiVersion: iam.cnrm.cloud.google.com/v1beta1
kind: IAMCustomRole
metadata:
  name: engineer
  namespace: dog-222
spec:
  permissions:
  - iap.tunnelInstances.accessViaIAP
  stage: GA
  title: Colorful Engineer
---
apiVersion: iam.cnrm.cloud.google.com/v1beta1
kind: IAMPolicyMember
metadata:
  name: iap-tunnel-blue
  namespace: dog-222
spec:
  member: user:blue@example.com
  resourceRef:
    apiVersion: resourcemanager.cnrm.cloud.google.com/v1beta1
    external: projects/dog-222
    kind: Project
  role: roles/engineer
---
apiVersion: iam.cnrm.cloud.google.com/v1beta1
kind: IAMPolicyMember
metadata:
  name: iap-tunnel-red
  namespace: dog-222
spec:
  member: user:red@example.com
  resourceRef:
    apiVersion: resourcemanager.cnrm.cloud.google.com/v1beta1
    external: projects/dog-222
    kind: Project
  role: roles/engineer
---
apiVersion: iam.cnrm.cloud.google.com/v1beta1
kind: IAMPolicyMember
metadata:
  name: iap-tunnel-yellow
  namespace: dog-222
spec:
  member: user:yellow@example.com
  resourceRef:
    apiVersion: resourcemanager.cnrm.cloud.google.com/v1beta1
    external: projects/dog-222
    kind: Project
  role: roles/engineer
EOF
```

In this output, the namespace of all instances is the
project name `dog-222`, and the project name also appears
in the resourceRef field of the `IAMPolicyMember` instances.

This project name only appears in the project _directory name_,
achieving our [DRY] goal.

Run the build:

<!-- @runIt @testAgainstLatestRelease -->
```
kustomize build $DEMO_HOME/projects/dog-222 >$DEMO_HOME/out_actual.yaml
```

and confirm that the actual output matches the expected output:

<!-- @diffShouldExitZero @testAgainstLatestRelease -->
```
diff $DEMO_HOME/out_actual.yaml $DEMO_HOME/out_expected.yaml
```

Build all the projects at once like this:

<!-- @buildAll @testAgainstLatestRelease -->
```
kustomize build $DEMO_HOME/all
```
