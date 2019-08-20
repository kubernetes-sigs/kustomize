# Generator Options

Kustomize provides options to modify the behavior of ConfigMap and Secret generators. These options include
 
 - disable appending a content hash suffix to the names of generated resources
 - adding labels to generated resources
 - adding annotations to generated resources
 
This demo shows how to use these options. First create a workspace.
```
DEMO_HOME=$(mktemp -d)
```

Create a kustomization and add a ConfigMap generator to it.

<!-- @createCMGenerator @testAgainstLatestRelease -->
```
cat > $DEMO_HOME/kustomization.yaml << EOF
configMapGenerator:
- name: my-configmap
  literals:	
  - foo=bar
  - baz=qux
EOF
```

Add following generatorOptions
<!-- @addGeneratorOptions @testAgainstLatestRelease -->
```
cat >> $DEMO_HOME/kustomization.yaml << EOF
generatorOptions:
 disableNameSuffixHash: true
 labels:
   kustomize.generated.resource: somevalue
 annotations:
   annotations.only.for.generated: othervalue
EOF
```
Run `kustomize build` and make sure that the generated ConfigMap
 
 - doesn't have name suffix
    <!-- @verify @testAgainstLatestRelease -->
    ```
    test 1 == \
    $(kustomize build $DEMO_HOME | grep "name: my-configmap$" | wc -l); \
    echo $?
    ```
 - has label `kustomize.generated.resource: somevalue`
     ```
     test 1 == \
     $(kustomize build $DEMO_HOME | grep -A 1 "labels" | grep "kustomize.generated.resource" | wc -l); \
     echo $?
     ```
 - has annotation `annotations.only.for.generated: othervalue`
      ```
      test 1 == \
      $(kustomize build $DEMO_HOME | grep -A 1 "annotations" | grep "annotations.only.for.generated" | wc -l); \
      echo $?
      ```
      
