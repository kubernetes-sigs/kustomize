# Exec plugin on linux in 60 seconds

This is a (no reading allowed!) 60 second copy/paste guided
example.  Full plugin docs [here](README.md).

This demo writes and uses a somewhat ridiculous
_exec_ plugin (written in bash) that generates a
`ConfigMap`.

This is a guide to try it without damaging your
current setup.

#### requirements

 * linux, git, curl, Go 1.12


## Make a place to work

```
DEMO=$(mktemp -d)
```

## Create a kustomization

Make a kustomization directory to
hold all your config:

```
MYAPP=$DEMO/myapp
mkdir -p $MYAPP
```

Make a deployment config:

```
cat <<'EOF' >$MYAPP/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: the-deployment
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: the-container
        image: monopole/hello:1
        command: ["/hello",
                  "--port=8080",
                  "--date=$(THE_DATE)",
                  "--enableRiskyFeature=$(ENABLE_RISKY)"]
        ports:
        - containerPort: 8080
        env:
        - name: THE_DATE
          valueFrom:
            configMapKeyRef:
              name: the-map
              key: today
        - name: ALT_GREETING
          valueFrom:
            configMapKeyRef:
              name: the-map
              key: altGreeting
        - name: ENABLE_RISKY
          valueFrom:
            configMapKeyRef:
              name: the-map
              key: enableRisky
EOF
```

Make a service config:

```
cat <<EOF >$MYAPP/service.yaml
kind: Service
apiVersion: v1
metadata:
  name: the-service
spec:
  type: LoadBalancer
  ports:
  - protocol: TCP
    port: 8666
    targetPort: 8080
EOF
```

Now make a config file for the plugin
you're about to write.

This config file is just another k8s resource
object.  The values of its `apiVersion` and `kind`
fields are used to _find_ the plugin code on your
filesystem (more on this later).

```
cat <<'EOF' >$MYAPP/cmGenerator.yaml
apiVersion: myDevOpsTeam
kind: SillyConfigMapGenerator
metadata:
  name: whatever
argsOneLiner: Bienvenue true
EOF
```

Finally, make a kustomization file
referencing all of the above:

```
cat <<EOF >$MYAPP/kustomization.yaml
commonLabels:
  app: hello
resources:
- deployment.yaml
- service.yaml
generators:
- cmGenerator.yaml
EOF
```

Review the files
```
ls -C1 $MYAPP
```


## Make a home for plugins

Plugins must live in a particular place for
kustomize to find them.

This demo will use the ephemeral directory:

```
PLUGIN_ROOT=$DEMO/kustomize/plugin
```

The plugin config defined above in
`$MYAPP/cmGenerator.yaml` specifies:

> ```
> apiVersion: myDevOpsTeam
> kind: SillyConfigMapGenerator
> ```

This means the plugin must live in a directory
named:

```
MY_PLUGIN_DIR=$PLUGIN_ROOT/myDevOpsTeam/sillyconfigmapgenerator

mkdir -p $MY_PLUGIN_DIR
```

The directory name is the plugin config's
_apiVersion_ followed by its lower-cased _kind_.

A plugin gets its own directory to hold itself,
its tests and any supplemental data files it
might need.

## Create the plugin

There are two kinds of plugins, _exec_ and _Go_.

Make an _exec_ plugin, installing it to the
correct directory and file name.  The file name
must match the plugin's _kind_ (in this case,
`SillyConfigMapGenerator`):

```
cat <<'EOF' >$MY_PLUGIN_DIR/SillyConfigMapGenerator
#!/bin/bash
# Skip the config file name argument.
shift
today=`date +%F`
echo "
kind: ConfigMap
apiVersion: v1
metadata:
  name: the-map
data:
  today: $today
  altGreeting: "$1"
  enableRisky: "$2"
"
EOF
```

By definition, an _exec_ plugin must be executable:

```
chmod a+x $MY_PLUGIN_DIR/SillyConfigMapGenerator
```

## Download kustomize 3.0.0

```
mkdir -p $DEMO/bin
gh=https://github.com/kubernetes-sigs/kustomize/releases/download
url=$gh/v3.0.0-pre/kustomize_3.0.0-pre_linux_amd64
curl -o $DEMO/bin/kustomize -L $url
chmod u+x $DEMO/bin/kustomize
```

## Review the layout

```
tree $DEMO
```

## Build your app, using the plugin:

```
XDG_CONFIG_HOME=$DEMO $DEMO/bin/kustomize build --enable_alpha_plugins $MYAPP
```

Above, if you had set

> ```
> PLUGIN_ROOT=$HOME/.config/kustomize/plugin
> ```

there would be no need to use `XDG_CONFIG_HOME` in the
_kustomize_ command above.

