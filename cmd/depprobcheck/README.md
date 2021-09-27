# Troubles with dependencies on openapi


If you see the error

> cannot use api.Schema.SchemaProps.Properties

then you have a set of incompatible dependencies.  This doc describes the problem and a fix.

tl;dr A mix of old and new is bad.

Anyone depending on k8s.io `v0.20.x` or _older_ packages must avoid depending on anything that depends
on `k8s.io/kube-openapi` newer than v0.0.0-20210323165736-1a6458611d18.

> More context in https://github.com/kubernetes/cli-runtime/issues/19

This dir exists to test the problem.

Edit the `main.go` and `go.mod` in this dir to see what builds
with various combinations of `cli-runtime` and `kube-openapi`.

####

A recent in change in kube-openapi

    https://github.com/kubernetes/kube-openapi/pull/234

means that anyone depending on

    k8s.io/cli-runtime@v0.20.4
   
and _any other package that imports kube-openapi_ (e.g. kyaml)
may see a build error like

>   ~/go/pkg/mod/sigs.k8s.io/kustomize@v2.0.3+incompatible/pkg/transformers/config/factorycrd.go:71:47:
>   cannot use api.Schema.SchemaProps.Properties (type map[string]"k8s.io/kube-openapi/pkg/validation/spec".Schema)
>   as type myProperties in argument to looksLikeAk8sType

## Why?

As it happens,

    k8s.io/cli-runtime@v0.20.4

depends on

    sigs.k8s.io/kustomize@v2.0.3+incompatible

Line 71 of factorycrd.go in kustomize v2.0.3 is:

	if !looksLikeAk8sType(api.Schema.SchemaProps.Properties) {

The `looksLikeAk8sType` function accepts the argument

    func looksLikeAk8sType(properties map[string]spec.Schema) bool {...}

At the call point in line 71 the argument is

    common.OpenAPIDefinition.Schema.SchemaProps.Properties

The file factorycrd.go depends on

    "github.com/go-openapi/spec"
    "k8s.io/kube-openapi/pkg/common"

The module sigs.k8s.io/kustomize@v2.0.3 predates Go modules.

To pin its dependencies, it has a Gopkg.lock file and
a vendor directory.  Per the lock file:

    sigs.k8s.io/kustomize@v2.0.3
    depends on "k8s.io/kube-openapi"
    revision = "b3f03f55328800731ce03a164b80973014ecd455"

Checking out this commit in the k8s.io/kube-openapi repo we see this

    k8s.io/kube-openapi/pkg/common/common.go:
    import  "github.com/go-openapi/spec"
    ...
    type OpenAPIDefinition struct {
      Schema       spec.Schema 
      Dependencies []string
    }

But per the imports in this file, `spec.Schema` lives in

    github.com/go-openapi/spec

The aforementioned Gopkg.lock file pins that at 

    sigs.k8s.io/kustomize@v2.0.3
    depends on "github.com/go-openapi/spec"
    revision = "bcff419492eeeb01f76e77d2ebc714dc97b607f5"

The struct is

    github.com/go-openapi/spec:
    type Schema struct {
      VendorExtensible
      SchemaProps
      SwaggerSchemaProps
      ExtraProps map[string]interface{} `json:"-"`
    }

    type SchemaProps struct {
      Properties map[string]Schema
    }

This is a recursive type; Schema holds a map[string]Schema.

All that is fine.

The problem arises when we build a binary that depends on both
kustomize v2.0.3 and, say, 

    k8s.io/kube-openapi v0.0.0-20210421082810-95288971da7e

This particular version of kube-openapi has a 'go.mod' file.

kube-openapi/pkg/common/common.go at tag "95288..." contains:

    import "k8s.io/kube-openapi/pkg/validation/spec"
    ...
    type OpenAPIDefinition struct {
      Schema       spec.Schema 
      Dependencies []string
    }

kube-openapi/pkg/validation/spec/schema.go at this tag contains:

    type Schema struct {
      VendorExtensible
      SchemaProps
      SwaggerSchemaProps
      ExtraProps map[string]interface{} `json:"-"`
   }

etc. etc. as above.  The same layout as above, but in different files.

So adding this new dependency means that 

    factorycrd.go 

is going to include a version of  "k8s.io/kube-openapi/pkg/common".
that deals in the type

    "k8s.io/kube-openapi/pkg/validation/spec".Schema

but will then attempt to pass this type to older code (the func
looksLikeAk8sType) which is looking to accept a

    "github.com/go-openapi/spec".Schema

It's the same type structure, but a different name so the 'linker' barfs.

To avoid this problem, one has to either

 * roll forward on cli-runtime -- depend on v0.21.0 or higher.
 
 * or stick with v0.20.4, which means retaining consistency
   with kustomize 2.0.3, which means depending on an older version
   of k8s.io/kube-openapi that still depends on go-openapi/spec.
 
   Fortunately you only have to go back to before PR
   https://github.com/kubernetes/kube-openapi/pull/234

   E.g. depend on k8s.io/kube-openapi v0.0.0-20210323165736-1a6458611d18

