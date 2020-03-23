# Install the Ambassador API Gateway

Ambassador is an API gateway technology built on top of Envoy with first-class Kubernetes integration. In this tutorial we'll go through the steps of setting up Ambassador on Kubernetes, and showing a brief example of use. The authoritative documentation on use and configuration will be on the [Ambassador website](https://www.getambassador.io/docs/).

## Installation and configuration

In this guide, we'll walk through the processing of deploying the Ambassador API Gateway in Kubernetes with [Kubernetes YAML](https://www.getambassador.io/docs/latest/topics/install/install-ambassador-oss/#kubernetes-yaml). You can also install it with [Helm](https://www.getambassador.io/docs/latest/topics/install/install-ambassador-oss/#helm).

### Deploying the Ambassador API Gateway

To deploy Ambassador in your default namespace, first you need to check if Kubernetes has RBAC enabled:

```bash
$ kubectl cluster-info dump --namespace default | grep authorization-mode
                            "--authorization-mode=Node,RBAC",
```

If you see something like `--authorization-mode=Node,RBAC` in the output, then RBAC is enabled. Then we just need to apply the ambassador yaml configuration:

```bash
$ kubectl apply -f ambassador-rbac.yaml
service/ambassador-admin created
clusterrole.rbac.authorization.k8s.io/ambassador created
serviceaccount/ambassador created
clusterrolebinding.rbac.authorization.k8s.io/ambassador created
deployment.apps/ambassador created
```

This creates a `Service`, a `ClusterRole`, a `ServiceAccount`, a `ClusterRoleBinding`, and a `Deployment`. `ClusterRoles` and `ClusterRoleBindings` are not namespaced. The other resources are created in the `default` namespace, you can specify the namespace to others such as `kube-system` which depends on your own demand. You can also download the latest version of the YAML file: 

```bash
$ wget https://www.getambassador.io/yaml/ambassador/ambassador-rbac.yaml
```

After installing, we can view the Ambassador diagnostic UI. By default, this is exposed to the internet at the URL `http://{{AMBASSADOR_HOST}}:8877/ambassador/v0/diag/`

### Defining the Ambassador Service

While the `Service` named `ambassador-admin` created previously is for providing an Ambassador Diagnostic web UI, we create a another `Service` that references the ambassador Deployment to expose it outside Kubernetes. It depends on your Kubernetes environment to specify the `Type of Service`, such as `LoadBalancer` and `NodePort`.

```bash
$ kubectl apply -f ambassador-service.yaml```
```

You can also expose the Ambassador deployment via [Host Network](https://www.getambassador.io/docs/latest/topics/install/bare-metal/#exposing-ambassador-via-host-network), the configuration in Deployment is commented in YAML file.

### Install Ambassador CRDs

Ambassador defines a number of [CRD resources](https://github.com/datawire/ambassador/blob/master/docs/reference/core/crds.md) in the `getambassador.io` API group for configuration usage. In order to use these `CRDs`, you should install them first:

```bash
$ kubectl apply -f ambassador-crds.yaml
```

## Using Ambassador with examples

The primary purpose of Ambassador API Gateway is to provide access and control to applications in Kubernetes. Here we use a demo to describe the configuration of routing process. The demo we create a Mapping configuration that tells Ambassador to route all traffic from /backend/ to the quote service with the following YAML file:

```bash
$ kubectl apply -f quote.yaml
```

In the YAML file, we define the `Mapping` resource to route the demo's traffic:
```bash
apiVersion: getambassador.io/v2
kind: Mapping
metadata:
  name: quote-backend
spec:
  prefix: /backend/
  service: quote
```

We can access the demo by typing the URL `http://{{AMBASSADOR_HOST}}:8080/backend/', you should see something similar to the following:
                                                                                          
```
{
    server: "tender-coconut-f4cccg84",
    quote: "Abstraction is ever present.",
    time: "2020-04-05T10:08:02.582881069Z"
}
```

In this tutorial you've learned how to set up Ambassador on Kubernetes. To go further, you can follow some of the guides on the ambassador website or try out a tool that builds on top of ambassador like Seldon.


## References

https://www.getambassador.io/docs/latest/topics/install/install-ambassador-oss/
https://www.getambassador.io/docs/latest/topics/install/bare-metal/
https://www.getambassador.io/docs/latest/topics/install/yaml-install/
https://gitlab.com/nibalizer/ambassador-iks/-/tree/master
