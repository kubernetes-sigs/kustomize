[builtin operations]: https://kubectl.docs.kubernetes.io/references/kustomize/builtins/
[builtin plugins]: https://kubectl.docs.kubernetes.io/references/kustomize/builtins/
[plugins]: https://kubectl.docs.kubernetes.io/references/kustomize/glossary/#plugin
[plugin]: https://kubectl.docs.kubernetes.io/references/kustomize/glossary/#plugin
[fields]: https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/
[fields in a kustomization file]: https://kubectl.docs.kubernetes.io/references/kustomize/kustomization/
[TransformerConfig]: ../api/internal/plugins/builtinconfig/transformerconfig.go
[kustomization]: https://kubectl.docs.kubernetes.io/references/kustomize/glossary/#kustomization

# Customizing kustomize

The [fields in a kustomization file] allow the user to
specify which resource files to use as input, how to
_generate_ new resources, and how to _transform_ those
resources - add labels, patch them, etc.

These fields are simple (low argument count) directives.
For example, the `commonAnnotations` field demands only a
list of _name:value_ pairs.

If using a field triggers behavior that pleases the user,
everyone's happy.

If not, the user can ask for new behavior to be implemented
in kustomize proper (and wait), or the user can write a
transformer or generator [plugin].  This latter option
generally means writing code; a Go plugin, a Go binary,
a C++ binary, a `bash` script - something.

There's a third option.  If one merely wants to tweak
behavior that already exists in kustomize, one may be able
to do so by just writing some YAML.

## Configure the builtin plugins

All of kustomize's [builtin operations] are implemented 
and usable as plugins.

Using the fields is convenient and brief, but necessarily
specifies only part of the entire plugin specification.  The
unspecified part is defaulted to what are hopefully
generally appealing values.

If, instead, one invokes the plugins directly using the
`transformers` or `generators` field, one can (indeed
_must_) specify the entire plugin configuration.

## Example: field vs plugin

Define a place to work:

<!-- @makeWorkplace @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)
```

### Using the `commonLabels` and `commonAnnotations` fields
 
In this simple example, we'll use just two resources: a deployment and a service.

Define them:

<!-- @makeRes1 @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deployment
spec:
  replicas: 10
  template:
    spec:
      containers:
      - name: the-container
        image: monopole/hello:1
EOF
```

<!-- @makeRes2 @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: service
spec:
  type: LoadBalancer
  ports:
  - protocol: TCP
    port: 8666
    targetPort: 8080
EOF
```

Now make a kustomization file that causes them
to be read and transformed:

<!-- @makeKustomization @testAgainstLatestRelease -->
```
cat <<'EOF' >$DEMO_HOME/kustomization.yaml
namePrefix: hello-
commonLabels:
  app: hello
commonAnnotations:
  area: "51"
  greeting: Take me to your leader
resources:
- deployment.yaml
- service.yaml
EOF
```

And run kustomize:

<!-- @checkLabel @testAgainstLatestRelease -->
```
kustomize build $DEMO_HOME
```

The output will be something like

> ```
> apiVersion: v1
> kind: Service
> metadata:
>   annotations:
>     area: "51"
>     greeting: Take me to your leader
>   labels:
>     app: hello
>   name: hello-service
> spec:
>   ports:
>   - port: 8666
>     protocol: TCP
>     targetPort: 8080
>   selector:
>     app: hello
>   type: LoadBalancer
> ---
> apiVersion: apps/v1
> kind: Deployment
> metadata:
>   annotations:
>     area: "51"
>     greeting: Take me to your leader
>   labels:
>     app: hello
>   name: hello-deployment
> spec:
>   replicas: 10
>   selector:
>     matchLabels:
>       app: hello
>   template:
>     metadata:
>       annotations:
>         area: "51"
>         greeting: Take me to your leader
>       labels:
>         app: hello
>     spec:
>       containers:
>       - image: monopole/hello:1
>         name: the-container
> ```

Let's say we are unhappy with this result.

We only want the annotations
to be applied down in the pod templates,
and don't want to see them in the metadata
for Service or Deployment.

We like that the label _app: hello_ ended up in

 - Service `spec.selector`
 - Deployment `spec.selector.matchLabels`
 - Deployment `spec.template.metadata.labels`

as this binds the Service (load balancer) to the pods,
and the Deployment itself to its own pods -
but we again don't care to see these labels in
the metadata for the Service and the Deployment.


### Configuring the builtin plugins instead

To fine tune this, invoke the same transformations
using the plugin approach.

Change the kustomization file:

<!-- @makeKustomization @testAgainstLatestRelease -->
```
cat <<'EOF' >$DEMO_HOME/kustomization.yaml
namePrefix: hello-
transformers:
- myAnnotator.yaml
- myLabeller.yaml
resources:
- deployment.yaml
- service.yaml
EOF
```

Then make the two plugin configuration files
(`myAnnotator.yaml`, `myLabeller.yaml`)
referred to by the `transformers` field above.
For details about the fields to specify, see
the documentation for the [builtin plugins].

<!-- @makeAnnotatorPluginConfig @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/myAnnotator.yaml
apiVersion: builtin
kind: AnnotationsTransformer
metadata:
  name: notImportantHere
annotations:
  area: 51
  greeting: take me to your leader
fieldSpecs:
- kind: Deployment
  path: spec/template/metadata/annotations
  create: true
EOF
```

<!-- @makeLabelPluginConfig @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/myLabeller.yaml
apiVersion: builtin
kind: LabelTransformer
metadata:
  name: notImportantHere
labels:
  app: hello
fieldSpecs:
- kind: Service
  path: spec/selector
  create: true
- kind: Deployment
  path: spec/selector/matchLabels
  create: true
- kind: Deployment
  path: spec/template/metadata/labels
  create: true
EOF
```

Finally, run kustomize again:

<!-- @checkLabel @testAgainstLatestRelease -->
```
kustomize build $DEMO_HOME
```

The output should resemble the following,
with fewer labels and annotations.

> ```
> apiVersion: v1
> kind: Service
> metadata:
>   name: hello-service
> spec:
>   ports:
>   - port: 8666
>     protocol: TCP
>     targetPort: 8080
>   selector:
>     app: hello
>   type: LoadBalancer
> ---
> apiVersion: apps/v1
> kind: Deployment
> metadata:
>   name: hello-deployment
> spec:
>   replicas: 10
>   selector:
>     matchLabels:
>       app: hello
>   template:
>     metadata:
>       annotations:
>         area: "51"
>         greeting: take me to your leader
>       labels:
>         app: hello
>     spec:
>       containers:
>       - image: monopole/hello:1
>         name: the-container
> ```


## The old way to do this

The original (and still functional) way to customize
kustomize is to specify a `configurations` field in the
kustomization file.

This field, normally omitted because it overrides hardcoded
data, accepts a list of file path arguments.  The files, in
turn, specify which fields in which k8s objects should be
affected by particular builtin transformations.  It's a
global configuration cutting across transformations, and
doesn't effect generators at all.

At `build` time, the configuration files are unmarshalled
into one instance of [TransformerConfig].  This
object, _plus_ the field values for `namePrefix`, etc.  are
fed into the transformation code to build the output.

The best way to write these custom configuration files is to
generate the files from the hardcoded values built into
kustomize via these commands:

> ```
> mkdir /tmp/junk
> kustomize config save -d /tmp/junk
> ```

One can then edit those file or files, and specify the
resulting edited files in a `configurations:`
field in a kustomization file used in a `build`.

Using plugins _completely ignores_ both hard coded
tranformer configuration, and any configuration loaded by
the `configuration` field.
