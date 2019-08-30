#TODO(howell): finish this

### tests:

running `kustomize build .` on without `--reorder` (or `--reorder=legacy`) uses
the legacy order. Expect Service, then Deployment, then Values
```
diff \
 <(printf "Service\nDeployment\nValues\n") \
 <(kustomize build . --enable_alpha_plugins | awk '/kind/ {print $2}')
 ```

running `kustomize build .` on with `--reorder=kubectlapply` uses the order
specified by the transformer configuration. Expect Service, then Values, then Deployment
```
diff \
 <(printf "Service\nValues\nDeployment\n") \
 <(kustomize build . --enable_alpha_plugins --reorder=kubectlapply | awk '/kind/ {print $2}')
 ```

running `kustomize build .` on with `--reorder=kubectldelete` uses the order
specified by the transformer configuration *in reverse*. Expect Deployment, then Values, then Service
```
diff \
 <(printf "Deployment\nValues\nService\n") \
 <(kustomize build . --enable_alpha_plugins --reorder=kubectldelete | awk '/kind/ {print $2}')
 ```
