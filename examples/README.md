English | [简体中文](zh/README.md)

# Examples

To run these examples, your `$PATH` must contain `kustomize`.
See the [installation instructions](../docs/INSTALL.md).

These examples are [tested](../travis/pre-commit.sh)
to work with the latest _released_ version of kustomize.

Basic Usage

  * [configGenerations](configGeneration.md) -
   Rolling update when ConfigMapGenerator changes.
   
  * [combineConfigs](combineConfigs.md) -
   Mixing configuration data from different owners
   (e.g. devops/SRE and developers).
  
  * [generatorOptions](generatorOptions.md) -
   Modifying behavior of all ConfigMap and Secret generators. 

  * [vars](wordpress/README.md) - Injecting k8s runtime data into
     container arguments (e.g. to point wordpress to a SQL service) by vars.
 
  * [image names and tags](image.md) - Updating image names and tags without applying a patch.
 
  * [remote target](remoteBuild.md) - Building a kustomization from a github URL
 
  * [json patch](jsonpatch.md) - Apply a json patch in a kustomization    

  * [patch multiple objects](patchMultipleObjects.md) - Apply a patch to multiple objects

Advanced Usage
  
- generator plugins:
  
   * [last mile helm](chart.md) - Make last mile modifications to
     a helm chart.
     
   * [secret generation](secretGeneratorPlugin.md) - Generating secrets from a plugin. 

- transformer plugins:
   * [validation transformer](validationTransformer/README.md) -
   validate resources through a transformer
  
- customize builtin transformer configurations
  
   * [transformer configs](transformerconfigs/README.md) - Customize transformer configurations
   

Multi Variant Examples

  * [hello world](helloWorld/README.md) - Deploy multiple
   (differently configured) variants of a simple Hello
   World server.
  
  * [LDAP](ldap/README.md) - Deploy multiple
     (differently configured) variants of a LDAP server.
     
  * [springboot](springboot/README.md) - Create a Spring Boot
   application production configuration from scratch.

  * [mySql](mySql/README.md) - Create a MySQL production
   configuration from scratch.
 
  * [breakfast](breakfast.md) - Customize breakfast for
     Alice and Bob.
     
  * [multibases](multibases/README.md) - Composing three variants (dev, staging, production) with a common base.  
