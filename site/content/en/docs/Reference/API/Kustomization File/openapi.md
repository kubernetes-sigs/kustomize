---
title: "openapi"
linkTitle: "openapi"
type: docs
weight: 14
description: >
    Specify where kustomize gets its OpenAPI schema.
---

Kustomize uses kubernetes OpenAPI data to get merge key and patch strategy 
information about resource types. Kustomize has an OpenAPI schema builtin, 
but this schema only has information about builtin kubernetes types. If
you need to provide merge key and patch strategy information about custom
resource types, you will have to provide your own OpenAPI schema to do so. 

In your kustomization file, you can specify where kustomize should get
its OpenAPI schema via an `openapi` field. For example:

```yaml
resources:
- my_resource.yaml

openapi:
  path: my_schema.json
```

The `openapi` field of a kustomization file can either a path to a custom schema
file, as in the example above. It can also be used to explicitly tell kustomize to
use a builtin kubernetes OpenAPI schema:

```yaml
resources:
- my_resource.yaml

openapi:
  version: v1.20.4
```

You can see what builtin kubernetes OpenAPI schemas are available with the command
`kustomize openapi info`. 

Here is an example of a custom resource we might want to edit with a custom OpenAPI schema
file. It looks like this: 

```yaml
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
```

This resource has an image field. Let's change its value from `server`
to `nginx` with a patch. You can get an OpenAPI document like this from
your locally favored cluster with the command `kustomize openapi fetch`.
Kustomize will use the OpenAPI extensions `x-kubernetes-patch-merge-key` and 
`x-kubernetes-patch-strategy` to perform a strategic merge. 
`x-kubernetes-patch-strategy` should be set to "merge", and you can set your 
merge key to whatever you like. 

Below, our custom resource inherits merge keys from PodTemplateSpec. In the
definition of "io.k8s.api.core.v1.Container", the `ports` field has its merge
key set to "containerPort":

```json
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
```

Then, our kustomization file to do the patch can be as follows:
```yaml
resources:
- my_resource.yaml

openapi:
  path: my_schema.json

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
```

