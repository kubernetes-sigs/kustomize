# kustomization of a helm chart

[`helm`]: https://helm.sh
[last mile]: https://testingclouds.wordpress.com/2018/07/20/844/
[artifact hub]: https://artifacthub.io
[_minecraft_]: https://artifacthub.io/packages/helm/minecraft-server-charts/minecraft
[plugin]: ../docs/plugins
[built]: https://kubectl.docs.kubernetes.io/references/kustomize/kustomization

Kustomize is [built] from _generators_ and
_transformers_; the former make kubernetes YAML, the
latter transform said YAML.

Kustomize, via the `helmCharts` field, has the ability to
use the [`helm`] command line program in a subprocess to
inflate a helm chart, generating YAML as part of (or as the
entirety of) a kustomize base.

This YAML can then be modified either in the base directly
(transformers always run _after_ generators), or via
a kustomize overlay.

Either approach can be viewed as [last mile] modification
of the chart output before applying it to a cluster.

The example below arbitrarily uses the
[_minecraft_] chart pulled from the [artifact hub]
chart repository.

## Preparation

This example defines the `helm` command as
<!-- @defineHelmCommand @testHelm -->
```
helmCommand=${MYGOBIN:-~/go/bin}/helmV3
```

This value is needed for testing this example in CI/CD.
A user doesn't need this if their binary is called
`helm` and is on their shell's `PATH`.


Make a place to work:

<!-- @makeWorkplace @testHelm -->
```
DEMO_HOME=$(mktemp -d)
mkdir -p $DEMO_HOME/base $DEMO_HOME/dev $DEMO_HOME/prod
```

## Define some variants

Define a kustomization representing your _development_
variant.

This could involve any number of kustomizations (see
other examples), but in this case just add the name
prefix '`dev-`' to all resources:

<!-- @writeKustDev @testHelm -->
```
cat <<'EOF' >$DEMO_HOME/dev/kustomization.yaml
namePrefix:  dev-
resources:
- ../base
EOF
```

Likewise define a _production_ variant, with a name
prefix '`prod-`':

<!-- @writeKustProd @testHelm -->
```
cat <<'EOF' >$DEMO_HOME/prod/kustomization.yaml
namePrefix:  prod-
resources:
- ../base
EOF
```

These two variants refer to a common base.

Define this base the usual way by creating a
`kustomization` file:

<!-- @writeKustBase @testHelm -->
```
cat <<'EOF' >$DEMO_HOME/base/kustomization.yaml
helmCharts:
- name: minecraft
  includeCRDs: false
  skipTests: true
  valuesInline:
    minecraftServer:
      eula: true
      difficulty: hard
      rcon:
        enabled: true
  releaseName: moria
  version: 3.1.3
  repo: https://itzg.github.io/minecraft-server-charts
EOF
```

The only thing in this particular file is a `helmCharts`
field, specifying a single chart.

The `valuesInline` field overrides some native chart values.

The `includeCRDs` field instructs Helm to generate 
`CustomResourceDefinitions`. 
See [the Helm documentation](https://helm.sh/docs/chart_best_practices/custom_resource_definitions/) for details.

Check the directory layout:

<!-- @tree -->
```
tree $DEMO_HOME
```

Expect something like:

> ```
> /tmp/whatever
> ├── base
> │  └── kustomization.yaml
> ├── dev
> │  └── kustomization.yaml
> └── prod
>    └── kustomization.yaml
> ```

### Helm related flags

Attempt to build the `base`:

<!-- @checkFailure @testHelm -->
```
cmd="kustomize build --helm-command $helmCommand $DEMO_HOME/base"
if ($cmd); then
   echo "Build should fail!" && false  # Force test to fail.
else
   echo "Build failed because no --enable-helm flag (desired outcome)."
fi
```

This `build` fails and complains about a missing
`--enable-helm` flag.

The flag `--enable-helm` exists to have the user
acknowledge that kustomize is running an external program as
part of the `build` step.  It's like the
`--enable-plugins` flag, but with a helm focus.

The flag `--helm-command` has a default value (`helm` of
course) so it's not suitable as an enablement flag.  A user
with `helm` on their `PATH` need not awkwardly specify
`'--helm-command helm'`.

Given the above, define a helper function to run `kustomize` with the
flags required for `helm` use in this demo:

<!-- @defineKustomizeIt @testHelm -->
```
function kustomizeIt {
  kustomize build \
    --enable-helm \
    --helm-command $helmCommand \
    $DEMO_HOME/$1
}
```
### Build the base and the variants

Now build the `base`:

<!-- @showBase @testHelm -->
```
kustomizeIt base
```

This works, and you see an inflated chart complete
with a `Secret`, `Service`, `Deployment`, etc.

As a side effect of this build, kustomize pulled the chart
and placed it in the `charts` subdirectory of the base.
Take a look:

<!-- @tree -->
```
tree $DEMO_HOME
```

If the chart had already been there, kustomize would
not have tried to pull it.

To change the location of the charts, use this
in your kustomization file:

> ```
> helmGlobals:
>  chartHome: charts
> ```

Change `charts` as desired, but it's best to keep it
in (or below) the same directory as the `kustomization.yaml` file.
If it's outside the kustomization root, the `build` command will
fail unless given the flag `'--load-restrictor=none'` to
disable file loading restrictions.

Now build the two variants `dev` and `prod`
and compare their differences:

<!-- @doCompare -->
```
diff <(kustomizeIt dev) <(kustomizeIt prod) | more
```

This shows so-called _last mile hydration_ of two variants
made from a common base that happens to be generated from a
helm chart.

## How does the pull work?

The command kustomize used to download the chart
is something like

> ```
> $helmCommand pull \
>    --untar \
>    --untardir $DEMO_HOME/base/charts \
>    --repo https://itzg.github.io/minecraft-server-charts \
>    --version 3.1.3 \
>    minecraft
> ```

The first use of kustomize above (when the `base` was
expanded) fetched the chart and placed it in the `charts`
directory next to the `kustomization.yaml` file.

This chart was reused, _not_ re-fetched, with the variant
expansions `prod` and `dev`.

If a chart exists, kustomize will not overwrite it (so to
suppress a pull, simply assure the chart is already in your
kustomization root).  kustomize won't check dates or version
numbers or do anything that smells like cache management.

> kustomize is a YAML manipulator.  It's not a manager
> of a cache of things downloaded from the internet.

## The pull happens once.

To show that the locally stored chart is being re-used, modify
its _values_ file.

First make note of the password encoded in the production
inflation:

<!-- @checkPassword @testHelm -->
```
test 1 == $(kustomizeIt prod | grep -c "rcon-password: Q0hBTkdFTUUh")
```

The above command succeeds if the value of the generated
password is as shown (`Q0hBTkdFTUUh`).

Now change the password in the local values file:

<!-- @valueChange @testHelm -->
```
values=$DEMO_HOME/base/charts/minecraft/values.yaml

grep CHANGEME $values
sed -i 's/CHANGEME/SOMETHING_ELSE/' $values
grep SOMETHING_ELSE $values
```

Run the build, and confirm that the same `rcon-password`
field in the output has a new value, confirming that the
chart used was a _local_ chart, not a chart freshly
downloaded from the internet:


<!-- @checkPassword2 @testHelm -->
```
test 1 == $(kustomizeIt prod | grep -c "rcon-password: U09NRVRISU5HX0VMU0Uh")
```

Finally, clean up:

<!-- @showBase @testHelm -->
```
rm -r $DEMO_HOME
```

## Performance

To recap, the helm-related kustomization fields make
kustomize run

> ```
> helm pull ...
> helm template ...
> ```

_as a convenience for the user_ to generate YAML from a helm chart.

Helm's `pull` command downloads the chart.  Helm's `template`
command inflates the chart template, spitting the inflated
template to stdout (where kustomize captures it) rather than
immediately sending it to a cluster as `helm install`
would.

To improve performance, a user can retain the chart after
the first pull, and commit the chart to their configuration
repository (below the `kustomization.yaml` file that refers
to it).  kustomize only tries to pull the chart if it's not
already there.

To further improve performance, a user can inflate the
chart themselves at the command line, e.g.

> ```
> helm template {releaseName} \
>     --values {valuesFile} \
>     --version {version} \
>     --repo {repo} \
>     {chartName} > {chartName}.yaml
> ```

then commit the resulting `{chartName}.yaml` file to a git
repo as a configuration base, mentioning that file as a
`resource` in a `kustomization.yaml` file, e.g.

> ```
> resources:
> - minecraft_v3.1.3_Chart.yaml
> ```

The user should choose when or if to refresh their local
copy of the chart's inflation.  kustomize would have no
awareness that the YAML was generated by helm, and kustomize
wouldn't run `helm` during the `build`.  This is analogous
to `Go` module vendoring.

### But it's not really about performance.

Although the `helm` related fields discussed above are handy
for experimentation and development, it's best to avoid them
in production.

The same argument applies to using _remote_ git URL's in
other kustomization fields.  Handy for experimentation,
but ill-advised in production.

It's irresponsible to depend on a remote configuration
that's _not under your control_.  Annoying enablement flags
like `'--enable-helm'` are intended to _remind_ one of a
risk, but offer zero protection from risk.  Further, they
are useless are reminders, since __annoying things are
immediately scripted away and forgotten__, as was done above
in the `kustomizeIt` shell function.

## Best practice

Don't use remote configuration that you don't control in
production.

Maintain a _local, inflated fork_ of a remote configuration,
and have a human rebase / reinflate that fork from time to
time to capture upstream changes.
