# kustomization of a helm chart

[last mile]: https://testingclouds.wordpress.com/2018/07/20/844/
[stable chart]: https://github.com/helm/charts/tree/master/stable
[helm charts]: https://github.com/helm/charts
[_minecraft_]: https://github.com/helm/charts/tree/master/stable/minecraft
[plugin]: ../docs/plugins

[Helm charts] aren't natively read by kustomize, but
kustomize has a builtin HelmChartInflationGenerator that allows one to
access helm charts.

One pattern combining kustomize and helm is
the [last mile] modification, where
one uses an inflated chart as a base, then
modifies it on the way to the cluster using
kustomize.

The example arbitrarily uses [_minecraft_] in [stable chart] repository,
but should work for any chart in any valid chart repository.

The following example assumes you have `helm`
on your `$PATH`. The plugin only supports helm V3 or later.

Make a place to work:

<!-- @makeWorkplace @helmtest -->

```
DEMO_HOME=$(mktemp -d)
mkdir -p $DEMO_HOME/base
mkdir -p $DEMO_HOME/dev
mkdir -p $DEMO_HOME/prod
```

## Use a remote chart

Define a kustomization representing your _development_
variant (aka environment).

This could involve any number of kustomizations (see
other examples), but in this case just add the name
prefix `dev-` to all resources:

<!-- @writeKustDev @helmtest -->

```
cat <<'EOF' >$DEMO_HOME/dev/kustomization.yaml
namePrefix:  dev-
resources:
- ../base
EOF
```

Likewise define a _production_ variant, with a name
prefix `prod-`:

<!-- @writeKustProd @helmtest -->

```
cat <<'EOF' >$DEMO_HOME/prod/kustomization.yaml
namePrefix:  prod-
resources:
- ../base
EOF
```

These two variants refer to a common base.

Define this base:

<!-- @writeKustDev @helmtest -->

```
cat <<'EOF' >$DEMO_HOME/base/kustomization.yaml
generators:
- chartInflator.yaml
EOF
```

The base refers to a generator configuration file
called `chartInflator.yaml`.

This file lets one specify the name of a [stable chart],
and other things like a path to a values file, defaulting
to the `values.yaml` that comes with the chart.

Create the config file `chartInflator.yaml`, specifying
the arbitrarily chosen chart name _minecraft_ and the repository
url that will be used to find the chart:

<!-- @writeGeneratorConfig @helmtest -->

```
cat <<'EOF' >$DEMO_HOME/base/chartInflator.yaml
apiVersion: builtin
kind: HelmChartInflationGenerator
metadata:
  name: notImportantHere
chartName: minecraft
chartRepoUrl: https://kubernetes-charts.storage.googleapis.com
EOF
```

Check the directory layout:

<!-- @tree -->

```
tree $DEMO_HOME
```

Expect something like:

> ```
> /tmp/whatever
> ├── base
> │   ├── chartInflator.yaml
> │   └── kustomization.yaml
> ├── dev
> │   └── kustomization.yaml
> └── prod
>    └── kustomization.yaml
> ```

Define a helper function to run kustomize with the
correct environment and flags for plugins:

<!-- @defineKustomizeIt @helmtest -->

```
function kustomizeIt {
  XDG_CONFIG_HOME=$DEMO_HOME \
  kustomize build --enable_alpha_plugins \
    $DEMO_HOME/$1
}
```

Finally, build the `prod` variant. Notice that all
resource names now have the `prod-` prefix:

<!-- @doProd @helmtest -->

```
clear
kustomizeIt prod
```

Compare `dev` to `prod`:

<!-- @doCompare -->

```
diff <(kustomizeIt dev) <(kustomizeIt prod) | more
```

To see the unmodified but inflated chart, run kustomize
on the base. Every invocation here is re-downloading
and re-inflating the chart.

<!-- @showBase @helmtest -->

```
kustomizeIt base
```

## Use a local chart

The example above fetches a new copy of the chart
and render it with each kustomize
build, because a local chart home isn't specified
in the configuration.

To suppress fetching, specify a _chart home_
explicitly, and just make sure the chart is already
there.

To demo this so that it won't interfere with your
existing helm environment, do this:

**This demo uses Helm V3**

<!-- @helmInit @helmtest -->

```
helmHome=$DEMO_HOME/dothelm
chartHome=$DEMO_HOME/base/charts
repoUrl=https://kubernetes-charts.storage.googleapis.com

function doHelm {
  helm --home $helmHome $@
}
```

Now download a chart; again use _minecraft_
(but you could use anything):

<!-- @fetchChart @helmtest -->

```
doHelm pull --untar \
    --untardir $chartHome \
    --repo $repoUrl \
    minecraft
```

The tree has more stuff now; helm config data
and a complete copy of the chart:

<!-- @tree -->

```
tree $DEMO_HOME
```

Add a `chartHome` field to the generator config file so
that it knows where to find the local chart:

<!-- @modifyGenConfig @helmtest -->

```
echo "chartHome: $chartHome" >>$DEMO_HOME/base/chartInflator.yaml
```

Change the values file, to show that the results
generated below are from the _locally_ stored chart:

<!-- @valueChange @helmtest -->

```
sed -i 's/CHANGEME!/SOMETHINGELSE/' $chartHome/minecraft/values.yaml
sed -i 's/LoadBalancer/NodePort/' $chartHome/minecraft/values.yaml
```

Finally, built it

<!-- @finalProd @helmtest -->

```
kustomizeIt prod
```

and observe the change from `LoadBalancer` to `NodePort`, and
the change in the encoded password.

Clean the directory.

<!-- @showBase @helmtest -->

```
rm -r $DEMO_HOME
```

## How to migrate from old plugin to builtin plugin

[bash-based helm chart inflator]: https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/someteam.example.com/v1/chartinflator
[go-based builtin helm chart inflator]: https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/builtin/helmchartinflationgenerator

The [bash-based helm chart inflator] is intended as an example of using bash
to write a generator plugin.

It proved to be popular for inflating helm charts,
so there's now a [go-based builtin helm chart inflator].

This newer generator is supported as part of the core code, so anyone using the
old bash-based example plugin would probably benefit from switching to the
newer built-in plugin.

Be advised that at the time of writing, the built-in plugin only supports helm v3.
