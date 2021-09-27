# Remote Sources

Kustomize supports building a [remote target], but the URLs are limited to common [Git repository specs].

To extend the supported format, Kustomize has a [plugin] system that allows one to integrate third-party tools such as [hashicorp/go-getter] to "download things from a string URL using a variety of protocols", extract the content and generated resources as part of kustomize build.

[remote target]: /examples/remoteBuild.md
[Git repository specs]: /api/internal/git/repospec_test.go
[plugin]: /docs/plugins
[hashicorp/go-getter]: https://github.com/hashicorp/go-getter

## Make a place to work

<!-- @makeWorkplace @test -->
```sh
DEMO_HOME=$(mktemp -d)
mkdir -p $DEMO_HOME/base
```

## Use a remote kustomize layer

Define a kustomization representing your _local_ variant (aka environment).

This could involve any number of kustomizations (see other examples), but in this case just add the name prefix `my-` to all resources:

<!-- @writeKustLocal @test -->
```sh
cat <<'EOF' >$DEMO_HOME/kustomization.yaml
namePrefix:  my-
resources:
- base/
EOF
```

It refer a remote base defined as below:

<!-- @writeKustLocal @test -->
```sh
cat <<'EOF' >$DEMO_HOME/base/kustomization.yaml
generators:
- goGetter.yaml
EOF
```

The base refers to a generator configuration file called `goGetter.yaml`.

This file lets one specify the source URL, and other things like sub path in the package, defaulting to base directory, and command to run under the path, defaulting to `kustomize build`.

Create the config file `goGetter.yaml`, specifying the arbitrarily chosen name _example_:

<!-- @writeGeneratorConfig @test -->
```sh
cat <<'EOF' >$DEMO_HOME/base/goGetter.yaml
apiVersion: someteam.example.com/v1
kind: GoGetter
metadata:
  name: example
url: github.com/kustless/kustomize-examples.git
EOF
```

Because this particular YAML file is listed in the `generators:` stanza of a kustomization file, it is treated as the binding between a generator plugin - identified by the _apiVersion_ and _kind_ fields - and other fields that configure the plugin.

Download the plugin to your `DEMO_HOME` and make it executable:

<!-- @installPlugin @test -->
```sh
plugin=plugin/someteam.example.com/v1/gogetter/GoGetter
curl -s --create-dirs -o \
"$DEMO_HOME/kustomize/$plugin" \
"https://raw.githubusercontent.com/\
kubernetes-sigs/kustomize/master/$plugin"

chmod a+x $DEMO_HOME/kustomize/$plugin
```

Define a helper function to run kustomize with the correct environment and flags for plugins:

<!-- @defineKustomizeIt @test -->
```sh
function kustomizeIt {
  XDG_CONFIG_HOME=$DEMO_HOME \
  kustomize build --enable_alpha_plugins \
    $DEMO_HOME/$1
}
```

Finally, build the local variant.  Notice that all
resource  names now have the `my-` prefix:

<!-- @doLocal @test -->
```sh
clear
kustomizeIt
```

Compare local variant to remote base:

<!-- @doCompare @test-->
```sh
diff <(kustomizeIt) <(kustomizeIt base) | more

...
<   name: my-remote-cm
---
>   name: remote-cm
```

To see the unmodified but extracted sources, run kustomize on the base.  Every invocation here is re-downloading and re-building the package.

<!-- @showBase @test -->
```sh
kustomizeIt base
```

## Use non-kustomize remote sources

Sometimes the remote sources does not include `kustomization.yaml`. To use that in the plugin, set command to override the default build.

<!-- @setCommand @test -->
```sh
echo "command: cat resources.yaml" >>$DEMO_HOME/base/goGetter.yaml
```

Finally, built it

<!-- @finalLocal @test -->
```sh
kustomizeIt
```

and observe the output includes raw `resources.yaml` instead of building result of remote `kustomization.yaml`.
