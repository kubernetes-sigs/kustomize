# Demo: Injecting k8s runtime data into containers

In this tutorial, you will learn how to use `kustomize` to declare a variable reference and substitute it in container's command.

To run WordPress, it's necessary to

- connect WordPress with a MySQL database
- access the service name of MySQL database from WordPress container

First make a place to work:
<!-- @makeDemoHome @test -->
```
DEMO_HOME=$(mktemp -d)
MYSQL_HOME=$DEMO_HOME/mysql
mkdir -p $MYSQL_HOME
WORDPRESS_HOME=$DEMO_HOME/wordpress
mkdir -p $WORDPRESS_HOME
```

### Download resources

Download the resources and `kustomization.yaml` for WordPress.

<!-- @downloadResources @test -->
```
CONTENT="https://raw.githubusercontent.com\
/kubernetes-sigs/kustomize\
/master/examples/wordpress/wordpress"

curl -s -o "$WORDPRESS_HOME/#1.yaml" \
  "$CONTENT/{deployment,service,kustomization}.yaml"
```

Download the resources and `kustomization.yaml` for MySQL.

<!-- @downloadResources @test -->
```
CONTENT="https://raw.githubusercontent.com\
/kubernetes-sigs/kustomize\
/master/examples/wordpress/mysql"

curl -s -o "$MYSQL_HOME/#1.yaml" \
  "$CONTENT/{deployment,service,secret,kustomization}.yaml"
```

### Create kustomization.yaml
Create a new kustomization with two bases:

<!-- @createKustomization @test -->
```
cat <<EOF >$DEMO_HOME/kustomization.yaml
bases:
  - wordpress
  - mysql
namePrefix: demo-
patches:
  - patch.yaml
EOF
```

### Download patch for WordPress
In the new kustomization, apply a patch for wordpress deployment. The patch does two things
- Add an initial container to show the mysql service name
- Add environment variable that allow wordpress to find the mysql database

<!-- @downloadPatch @test -->
```
CONTENT="https://raw.githubusercontent.com\
/kubernetes-sigs/kustomize\
/master/examples/wordpress"

curl -s -o "$DEMO_HOME/#1.yaml" \
  "$CONTENT/{patch}.yaml"
```
The patch has following content
> ```
> apiVersion: apps/v1beta2
> kind: Deployment
> metadata:
>   name: wordpress
> spec:
>   template:
>     spec:
>       initContainers:
>       - name: init-command
>         image: debian
>         command:
>         - "echo $(WORDPRESS_SERVICE)"
>         - "echo $(MYSQL_SERVICE)"
>       containers:
>       - name: wordpress
>         env:
>         - name: WORDPRESS_DB_HOST
>           value: $(MYSQL_SERVICE)
>         - name: WORDPRESS_DB_PASSWORD
>           valueFrom:
>             secretKeyRef:
>               name: mysql-pass
>               key: password
> ```
The init container's command requires information that depends on k8s resource object fields, represented by the placeholder variables
$(WORDPRESS_SERVICE) and $(MYSQL_SERVICE).

### Bind the Variables to k8s Object Fields

<!-- @addVarRef @test -->
```
cat <<EOF >>$DEMO_HOME/kustomization.yaml
vars:
  - name: WORDPRESS_SERVICE
    objref:
      kind: Service
      name: wordpress
      apiVersion: v1
    fieldref:
      fieldpath: metadata.name
  - name: MYSQL_SERVICE
    objref:
      kind: Service
      name: mysql
      apiVersion: v1
EOF
```
`WORDPRESS_SERVICE` is from the field `metadata.name` of Service `wordpress`. If we don't specify `fieldref`, the default is `metadata.name`. So `MYSQL_SERVICE` is from the field `metadata.name` of Service `mysql`.

### Substitution
Confirm the variable substitution:

<!-- @kustomizeBuild @test -->
```
kustomize build $DEMO_HOME
```

Expect this in the output:

> ```
> (truncated)
> ...
>      initContainers:
>      - command:
>        - echo demo-wordpress
>        - echo demo-mysql
>        image: debian
>        name: init-command
>
> ```