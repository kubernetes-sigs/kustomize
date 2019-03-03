# RemoteKubernetesSecret

`remoteKubernetesSecret` allows you to source your secrets from existing secrets on a Kubernetes cluster which you have access to.

To try this example you can create a cluster with [minikube][0] and ensure that [kubectl][1] is configured for it as the [current context][2]:

```shell
minikube start
```

(See the [minikube documentation][0] for further assistance getting it set up for your environment)

Once everything is working, run the following to create some "mock" secrets which we'll use for the example:

```shell
echo 'test1' > /tmp/kustomize-remote-k8s-test-1
echo 'test2' > /tmp/kustomize-remote-k8s-test-2
```

And then store them in `minikube`:

```shell
kubectl create namespace mynamespace
kubectl -n mynamespace create secret generic mysecret \
    --from-file=mykey1=/tmp/kustomize-remote-k8s-test-1 \
    --from-file=mykey2=/tmp/kustomize-remote-k8s-test-2
```

Then run this example with:

```shell
kustomize build examples/remoteKubernetesSecretSource/
```

To cleanup `minikube`, you can run:

```shell
kubectl -n mynamespace delete secret mysecret
```

Optionally (assuming you're not otherwise going to be using the namespace) you can remove the created namespace entirely:

```shell
kubectl delete namespace mynamespace
```

[0]:https://kubernetes.io/docs/tasks/tools/install-minikube/
[1]:https://kubernetes.io/docs/tasks/tools/install-kubectl/
[2]:https://kubernetes.io/docs/tasks/access-application-cluster/configure-access
