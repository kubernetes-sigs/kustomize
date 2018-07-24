# Examples

These examples assume that `kustomize` is on your `$PATH`.

They are covered by [pre-commit](../bin/pre-commit.sh)
tests, and should work with HEAD

<!-- @installkustomize @test -->
```
go get github.com/kubernetes-sigs/kustomize
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

 * [breakfast](breakfast.md) - Customize breakfast for
   Alice and Bob.
   
 * [container args](wordpress/README.md) - Injecting k8s runtime data into container arguments (e.g. to point wordpress to a SQL service).
 
 * [image tags](imageTags.md) - Updating image tags without applying a patch.
