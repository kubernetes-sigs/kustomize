# Using a Custom OpenAPI schema

For more details regarding this feature you can read the
[Kustomize OpenAPI Features KEP](https://github.com/kubernetes/enhancements/tree/master/keps/sig-cli/2206-openapi-features-in-kustomize).

A kustomization file supports adding your own
OpenAPI schema to define merge keys and patch
strategy.

Make a place to work:

<!-- @placeToWork @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)
```

We'll be editing our own [custom resource](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/).

<!-- @customOpenAPI @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/my_resource.yaml
apiVersion: example.com/v1alpha1
kind: MyResource
metadata:
  name: service
spec:
  template:
    spec:
      containers:
      - name: server
        image: server
        command: example
        ports:
        - name: grpc
          protocol: TCP
          containerPort: 8080
EOF
```

This resource has an image field. Let's change its value from server 
to nginx with a patch. 

Kustomize gets its merge key information from the OpenAPI data
provided by the kubernetes API server. It doesn't have information
about custom resources, so we will have to provide our own 
schema file. 

Note: CRDs support declarative validation using an OpenAPI v3 schema.
See https://book.kubebuilder.io/reference/generating-crd.html#validation.

You can get an OpenAPI document like this by fetching the OpenAPI
document from your locally favored cluster with the command
`kustomize openapi fetch`. Kustomize will use the OpenAPI extensions
`x-kubernetes-patch-merge-key` and `x-kubernetes-patch-strategy` to
perform a strategic merge. `x-kubernetes-patch-strategy` should be set
to "merge", and you can set your merge key to whatever you like. Below,
our custom resource inherits merge keys from `PodTemplateSpec`. 

<!-- @addCustomSchema @testAgainstLatestRelease -->
```
cat <<EOF >>$DEMO_HOME/mycr_schema.json
{
  "definitions": {
    "v1alpha1.MyResource": {
      "properties": {
        "apiVersion": {
          "type": "string"
        },
        "kind": {
          "type": "string"
        },
        "metadata": {
          "type": "object"
        },
        "spec": {
          "properties": {
            "template": {
              "\$ref": "#/definitions/io.k8s.api.core.v1.PodTemplateSpec"
            }
          },
          "type": "object"
        },
        "status": {
           "properties": {
            "success": {
              "type": "boolean"
            }
          },
          "type": "object"
        }
      },
      "type": "object",
      "x-kubernetes-group-version-kind": [
        {
          "group": "example.com",
          "kind": "MyResource",
          "version": "v1alpha1"
        }
      ]
    },
    "io.k8s.api.core.v1.PodTemplateSpec": {
      "properties": {
        "metadata": {
          "\$ref": "#/definitions/io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta"
        },
        "spec": {
          "\$ref": "#/definitions/io.k8s.api.core.v1.PodSpec"
        }
      },
      "type": "object"
    },
    "io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta": {
      "properties": {
        "name": {
          "type": "string"
        }
      },
      "type": "object"
    },
    "io.k8s.api.core.v1.PodSpec": {
      "properties": {
        "containers": {
          "items": {
            "\$ref": "#/definitions/io.k8s.api.core.v1.Container"
          },
          "type": "array",
          "x-kubernetes-patch-merge-key": "name",
          "x-kubernetes-patch-strategy": "merge"
        }
      },
      "type": "object"
    },
    "io.k8s.api.core.v1.Container": {
      "properties": {
        "command": {
          "items": {
            "type": "string"
          },
          "type": "array"
        },
        "image": {
          "type": "string"
        },
        "name": {
          "type": "string"
        },
        "ports": {
          "items": {
            "\$ref": "#/definitions/io.k8s.api.core.v1.ContainerPort"
          },
          "type": "array",
          "x-kubernetes-list-map-keys": [
            "containerPort",
            "protocol"
          ],
          "x-kubernetes-list-type": "map",
          "x-kubernetes-patch-merge-key": "containerPort",
          "x-kubernetes-patch-strategy": "merge"
        }
      },
      "type": "object"
    },
    "io.k8s.api.core.v1.ContainerPort": {
      "properties": {
        "containerPort": {
          "format": "int32",
          "type": "integer"
        },
        "name": {
          "type": "string"
        },
        "protocol": {
          "type": "string"
        }
      },
      "type": "object"
    }
  }
}

EOF
```

We'll of course need a `kustomization` file
referring to the custom resource, and containing our patch:

<!-- @openAPIkustomization @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/kustomization.yaml
resources:
- my_resource.yaml

openapi:
  path: mycr_schema.json

patchesStrategicMerge:
- |-
  apiVersion: example.com/v1alpha1
  kind: MyResource
  metadata:
    name: service
  spec:
    template:
      spec:
        containers:
        - name: server
          image: nginx
EOF
```

Define the expected output:
<!-- @expected @testAgainstLatestRelease -->
```
cat <<EOF >$DEMO_HOME/out_expected.yaml
apiVersion: example.com/v1alpha1
kind: MyResource
metadata:
  name: service
spec:
  template:
    spec:
      containers:
      - command: example
        image: nginx
        name: server
        ports:
        - containerPort: 8080
          name: grpc
          protocol: TCP
EOF
```

Run the build:
<!-- @runExample @testAgainstLatestRelease -->
```
kustomize build $DEMO_HOME >$DEMO_HOME/out_actual.yaml
```

Confirm they match:

<!-- @diffShouldBeEmpty @testAgainstLatestRelease -->
```
diff $DEMO_HOME/out_actual.yaml $DEMO_HOME/out_expected.yaml
```
