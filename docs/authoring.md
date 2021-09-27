# kustomization authoring

kustomize provides sub-commands for managing the contents of a kustomization file from the command line.

## kustomize create

The `kustomize create` command will create a new kustomization in the current directory.

When run without any flags the command will create an empty `kustomization.yaml` file that can then be updated manually or with the `kustomize edit` sub-commands.

```
kustomize create --namespace=myapp --resources=deployment.yaml,service.yaml --label=app=myapp
```

### Detecting resources

> NOTE: Resource detection will not follow symlinks.

Flags:

      --annotation string   Add one or more common annotations.
      --autodetect          Search for kubernetes resources in the current directory to be added to the kustomization file.
      --help, -h            help for create
      --label string        Add one or more common labels.
      --nameprefix string   Sets the value of the namePrefix field in the kustomization file.
      --namespace string    Set the value of the namespace field in the customization file.
      --namesuffix string   Sets the value of the nameSuffix field in the kustomization file.
      --recursive           Enable recursive directory searching for resource auto-detection.
      --resources string    Name of a file containing a file to add to the kustomization file.

## kustomize edit

With an existing kustomization file the `kustomize edit` command 

* add
* set
* remove
* fix
