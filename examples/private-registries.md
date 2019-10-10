# Demo: Setting up a private docker registry

In order to load images from a private image registry you will need to setup a kubernetes secret with token and user info for that registry and then refer to it in your deployment `imagePullSecrets`


<!-- @createIngress @testAgainstLatestRelease -->
```
DEMO_HOME=$(mktemp -d)

As we dont want anyone else off of our machine to have our credentials, ignore the imagePullSecrets.txt

cat <<EOF >$DEMO_HOME/.gitignore
imagePullSecret.txt
```

Next create your kustomization.yaml and refer to a secret generator

cat <<EOF >$DEMO_HOME/kustomization.yaml
namespace: development

secretGenerator:
  - name: demo-app-regcred
    files:
      - .dockerconfigjson=imagePullSecret.txt
    type: kubernetes.io/dockerconfigjson
EOF

Then enter your details here.

cat <<EOF >$DEMO_HOME/imagePullSecret.txt
{
    "auths": {
        "$REGISTRY_ADDRESS_HERE": {
            "username": "$USER",
            "password": "$PASSWORD",
            "email": "$EMAIL"
        }
    }
}
EOF

cat <<EOF >$DEMO_HOME/app-k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo-app-name
spec:
  ...
  template:
    ...
    spec:
      imagePullSecrets:
        - name: demo-app-regcred
EOF


### Note

Obviously the credentials will then be installed into the cluster that you install it on, try not to expose your personal creds.