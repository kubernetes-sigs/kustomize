### Demo

Please sure you are under the path `examle/deployment directory`.

### Environment

* kustomize version: v5.0.1

* Render the base resources:

```bash
$ kustomize build base
```

* Render the test resources:

```bash
$ kustomize build overlays/test/
```

* Render the production resources:

```bash
$ kustomize build overlays/production/ 
```
