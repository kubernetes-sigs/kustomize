# Examples

These examples assume that `kustomize` is on your `$PATH`.

They are covered by [pre-commit](../bin/pre-commit.sh)
tests, and should work with HEAD

<!-- @installkustomize @test -->
```
go get sigs.k8s.io/kustomize
```

 * [hello world](helloWorld/README.md) - Deploy multiple
   (differently configured) variants of a simple Hello
   World server.

 * [LDAP](ldap/README.md) - Deploy multiple
   (differently configured) variants of a LDAP server.

 * [mySql](mySql/README.md) - Create a MySQL production
   configuration from scratch.

 * [springboot](springboot/README.md) - Create a Spring Boot
   application production configuration from scratch.

 * [combineConfigs](combineConfigs.md) -
   Mixing configuration data from different owners
   (e.g. devops/SRE and developers).
   
 * [configGenerations](configGeneration.md) -
   Rolling update when ConfigMapGenerator changes
 
 * [generatorOptions](generatorOptions.md) - Modifying behavior of all ConfigMap and Secret generators.  

 * [breakfast](breakfast.md) - Customize breakfast for
   Alice and Bob.
   
 * [vars](wordpress/README.md) - Injecting k8s runtime data into container arguments (e.g. to point wordpress to a SQL service) by vars.
 
 * [image names and tags](image.md) - Updating image names and tags without applying a patch.

 * [multibases](multibases/README.md) - Composing three variants (dev, staging, production) with a common base.

 * [remote target](remoteBuild.md) - Building a kustomization from a github URL
 
 * [json patch](jsonpatch.md) - Apply a json patch in a kustomization
 
 * [transformer configs](transformerconfigs/README.md) - Customize transformer configurations
 
