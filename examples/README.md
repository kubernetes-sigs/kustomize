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

 * [configGeneration](configGeneration.md) -
   Mixing configuration data from different owners
   (e.g. devops/SRE and developers).

 * [breakfast](breakfast.md) - Customize breakfast for
   Alice and Bob.
