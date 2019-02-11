# The VarnishService Kubernetes Operator

[![Build Status](https://wcp-twc-icmkube-jenkins.swg-devops.com/job/TheWeatherCompany%20ICM/job/icm-varnish-k8s-operator/job/master/badge/icon)](https://wcp-twc-icmkube-jenkins.swg-devops.com/job/TheWeatherCompany%20ICM/job/icm-varnish-k8s-operator/job/master/)

VarnishService fills a space currently missing within Kubernetes on IBM Cloud: Varnish. IBM does not provide any managed Varnish instances, and Kubernetes does not have anything that works like Varnish does. Thus, this project aims to fill that space by providing a convenient way to deploy Varnish instances.

By default, deploying a Varnish directly as a Deployment into Kubernetes is not immediately useful because the VCL must have IP addresses for its backends. The only obvious way to get an IP address is via a Kubernetes Service, but that Service already acts as a load balancer to the Deployment it backs, which means undefined behavior from the Varnish perspective, and adds an extra network hop. Thus, trying to use Varnish in a regular deployment is unproductive.

Instead, the VarnishService operator manages the deployment of your Varnish, filling in the IP addresses of the pods for you, and manages the required infrastructure. The operator itself is made up of 2 components:

**CustomResourceDefinition**: the actual "VarnishService", that acts in the same way that a Service resource does, except with an added Varnish layer between the Service and the Deployment it backs. You would define a resource of Kind "VarnishService", and specify all the regular specs for a Service, plus some new fields that control how many Varnish instances you want, how much memory/cpu they get, and other relevant information for the Varnish cluster.

**Controller**: The controller is an application deployed into your cluster that knows how to react to the VarnishService CustomResource. Meaning, this application watches for new or changed VarnishServices and handles the actual underlying infrastructure. That means it must be running at all times in the cluster, although it lives in its own namespace away from your application.

## Kubernetes Version Requirement

This operator assumes that the `/status` and `/scale` subresources are enabled for Custom Resources, which means that you must have enabled this alpha feature for Kubernetes v1.10 (impossible on IBM Kubernetes Service) or are using at least v1.11, where it is enabled by default.

## Installation

The VarnishService Operator is packaged as a [Helm Chart](https://helm.sh/), hosted on [Artifactory](https://na.artifactory.swg-devops.com). To get access to this Artifactory, you must be a user on the Weather Channel Bluemix account 1638245.

### Getting Helm Access

After you are a user on the correct Bluemix account, you must generate an API key within [Artifactory](https://na.artifactory.swg-devops.com) for Helm to use. You can generate an API key on your profile page, found in the upper-right of the home page. Using that generated API Key, you can log in to Helm using [these instructions](https://www.jfrog.com/confluence/display/RTF/Helm+Chart+Repositories), where the username is your email and the password is your API key. Specifically, that will look like:

```sh
helm repo add wcp-icm-helm-virtual https://na.artifactory.swg-devops.com/artifactory/wcp-icm-helm-virtual --username=<your-email> --password=<encrypted-password>
helm repo update
```

### Getting Container Registry Access

As part of the helm install, you will also need access to the Container Registry in order to pull the Docker images associated with the Helm charts. This can be done using the IBMCloud CLI:

```sh
ibmcloud cr token-add --non-expiring --description 'for Varnish operator'
```

And from the output, save the `Token` field.

### Adding The Key To The Namespace

Once you have generated your docker registry key, you must either use an existing or create a new namespace. Add a secret with the docker registry token to that namespace:

```sh
kubectl create secret docker-registry <name> --namespace <namespace> --docker-server=registry.ng.bluemix.net --docker-username=token --docker-password=<token> --docker-email=<any-email>
```

Note that

* `<name>` can be any name, e.g. `docker-reg-secret`
* `docker-username` MUST be `token`
* `docker-email` can be any email. For example, `a@b.c`

By default the Helm install will assume a namespace called `varnish-operator-system` exists.

### Configuring The Operator

The operator has options to customize the installation into your cluster, exposed as values in the Helm `values.yaml` file. [See the default `values.yaml` annotated with descriptions of each field](/varnish-operator/values.yaml) to see what can be customized when deploying this operator.

### Installing The Operator

Once a Namespace has been created with a docker registry secret and an appropriate `values.yaml` has been assembled, install the operator using

```sh
helm upgrade --install <name-of-release> wcp-icm-helm-virtual/varnish-operator --version <latest-version> --wait --namespace <namespace-with-registry-token>
```

Note that

* `<name-of-release>` can be any name and has the same meaning as `<name>` for `helm install --name <name>`
* `<namespace-with-registry-token>` must match `namespace` in the `values.yaml` file.

## Usage

Once the operator is installed, your cluster should have a new resource, `varnishservice` (with aliases `varnishservices` and `vs`). From this point, you can create a yaml file with the `VarnishService` Kind.

### Configuring Access

Since the VarnishService requires pulling images from the same private repository as the Operator, the same docker registry key must exist in the target namespace for the VarnishService. Thus, add a secret with the docker registry token to that namespace before creating the resource.

### Configuring The VarnishService Resource

VarnishService has [an example yaml file annotated with descriptions of each field](/config/samples/icm_v1alpha1_varnishservice.yaml) To see what can be customized for the VarnishService.

### Preparing VCL Code

There are 3 fields relevant to configuring the VarnishService for VCL code, in `spec.vclConfigMap`:

* **name**: This is a REQUIRED field, and tells the VarnishService the name of the ConfigMap that contains/will contain the VCL files
* **backendsFile**: The name of the file that will contain VCL regarding backends. To be exact, the VarnishService will expect to see a `<backendsFile>.tmpl` file in the ConfigMap that contains the Go template to be used to generate the `<backendsFile>`. For example, if `backendsFile=backends.vcl`, there should be a `backends.vcl.tmpl` file in the ConfigMap
* **defaultFile**: The name of the file that acts as the entrypoint for Varnish. This is the name of the file that will be passed to the Varnish executable

Beyond the `backendsFile` template and the `defaultFile`, you can place any other VCL files in the ConfigMap and they will land in the same folder as the aforementioned files.

If a ConfigMap of name `spec.vclConfigMap.name` does not exist on VarnishService creation, the operator will create one and populate it with a default `<backendsFile>.tmpl` and `<defaultFile>`. Their behavior are as follows:

* [`<backendsFile>.tmpl`](/config/vcl/backends.vcl.tmpl): collect all backends into a single director and round-robin between them
* [`<defaultFile>`](/config/vcl/default.vcl):
  * respond to `GET /heartbeat` checks with a 200
  * respond to `GET /liveness` checks with a 200 or 503, depending on healthy backends
  * respond to all other requests normally, caching all non-404 responses
  * hash request based on url
  * add `X-Varnish-Cache` header to response with "HIT" or "MISS" value, based on presence in cache

If you would like to use the default `<backendsFile>.tmpl`, but a custom `<defaultFile>`, the easiest way is to create the VarnishService without the ConfigMap, let the operator create the ConfigMap for you, and then modify the contents of the ConfigMap after creation. Alternatively, just copy the content as linked above.

### Using user defined VCL Code versions

VCL related status information is available at field `vcl` in status object. 

The current VCl version can be found in the `vcl.configMapVersion` status field. It matches the resource version of the config map that contains the VCL code. 

For user readable versions an annotation `VCLVersion` can be used. It should be set for the config map where the VCL configuration is defined.

```bash
> kubectl -n varnish-ns get cm varnish-config -o yaml
apiVersion: v1
kind: ConfigMap
metadata:
  annotations:
    VCLVersion: v1.0 # <-- set by the user
  creationTimestamp: "2018-12-21T12:59:07Z"
  resourceVersion: "292181"
    ...
data:
    ...
```

After setting the annotation, that version can be seen in the status field `vcl.version` of the varnish service. This field is optional and not present if the version is not set in the config map annotation.

```bash
> kubectl -n varnish-ns get vs my-varnish -o yaml
apiVersion: icm.ibm.com/v1alpha1
kind: VarnishService
metadata:
    ...
status:
  vcl:
    version: v1.0 # <-- reflects the `VCLVersion` annotation in the config map
    configMapVersion: "292181" # <-- reflects the config map resource version
  deployment:
    affinity:
      podAntiAffinity:
    ...   
```

After the VCL in the config map has been changed, the status field will be immediately updated to reflect the latest version. However that does not guarantee that Varnish pods run the latest VCL configuration. It needs time to reload and even could fail to reload if the VCL has syntax error for example.
 
To give users a better observability about currently running VCL versions the status has a field `vcl.availability` which indicates how many pods have the latest version and how many of them are outdated. 

```bash
> kubectl -n varnish-ns get vs my-varnish -o yaml
apiVersion: icm.ibm.com/v1alpha1
kind: VarnishService
metadata:
  annotations:
    ...
status:
  vcl:
    configMapVersion: "292181"
    version: v1.0
    availability: 1 latest / 0 outdated # <-- all pods have the latest VCL version
  deployment:
    availableReplicas: 1
    conditions:
    ...
```

To check which pods have outdated versions, simply check their annotations. The annotation `configMapVersion` on the Varnish pod will indicate the latest version of the config map used. If it's not the same as in the VarnishService status it's likely that there's an issue.

Example of detecting a pod that failed to reload:

```bash
# get the latest version
> kubectl get varnishservice -n varnish-ns my-varnish -o=custom-columns=NAME:.metadata.name,CONFIG_MAP_VERSION:.status.vcl.configMapVersion
NAME        CONFIG_MAP_VERSION
my-varnish  292181
# figure out which pods doesn't have that latest version
> kubectl get pods -n varnish-ns -o=custom-columns=NAME:.metadata.name,CONFIG_MAP_VERSION:.metadata.annotations.configMapVersion
NAME                                            CONFIG_MAP_VERSION
my-varnish-varnish-deployment-545f475b58-7xn9k  292181
my-varnish-varnish-deployment-545f475b58-jc5vg  292181
my-varnish-varnish-deployment-545f475b58-nqqd2  351231 #outdated VCL
# check logs for that pod with outdated VCL
> kubectl logs -n my-varnish my-varnish-varnish-deployment-545f475b58-nqqd2 
2018-12-21T17:03:07.917Z	INFO	controller/controller.go:124	Rewriting file	{"path": "/etc/varnish/backends.vcl"}
2018-12-21T17:03:17.904Z	ERROR	controller/controller.go:157	exit status 1
/go/src/icm-varnish/k-watcher/pkg/controller/controller_varnish.go:50: Message from VCC-compiler:
Expected one of
	'acl', 'sub', 'backend', 'probe', 'import', 'vcl',  or 'default'
Found: 'dsafdf' at
('/etc/varnish/backends.vcl' Line 4 Pos 2)
 dsafdf
-######

Running VCC-compiler failed, exited with 2
Command failed with error code 106
VCL compilation failed
No VCL named v304255 known.
Command failed with error code 106

/go/src/icm-varnish/k-watcher/vendor/sigs.k8s.io/controller-runtime/pkg/internal/controller/controller.go:207: 
icm-varnish/k-watcher/pkg/logger.WrappedError
	/go/src/icm-varnish/k-watcher/pkg/logger/logger.go:49
ic
```

As the logs indicate, the issue here is the invalid VCL syntax.

### Creating a VarnishService Resource

Once the VarnishService resource yaml is ready, simply `kubectl apply -f <varnish-service>.yaml` to create the resource. Once complete, you should see:

* a deployment with the name `<varnish-service-name>-deployment`. This is the Varnish cluster, and should have inherited everything under the `deployment` part of the spec.
* 2 services, one `<varnish-service-name>-cached` and one `<varnish-service-name>-nocached`. As is implied by the names, using `<varnish-service-name>-cached` will direct to Varnish, which then forwards to the underlying deployment, while `<varnish-service-name>-nocached` will target the underlying deployment directly, with no Varnish caching. `<varnish-service-name>-cached` will have inherited everything under the `service` part of the spec, other than its `selector` and `port`, which will be redirected to the Varnish deployment.
* A ConfigMap with VCL in it (either user-created, before running `kubectl apply -f <varnish-service>.yaml`, or generated by operator, after)
* A role/rolebinding/clusterrole/clusterrolebinding/serviceAccount combination to give the Varnish deployment the ability to access necessary resources.

### Updating a VarnishService Resource

Just as with any other Kubernetes resource, using `kubectl apply`, `kubectl patch`, or `kubectl replace` will all update the VarnishService appropriately. The operator will handle how that update propagates to its dependent resources.

### Deleting a VarnishService Resource

Simply calling `kubectl delete` on the VarnishResource will recursively delete all dependent resources, so that is the only action you need to take. This includes a user-generated ConfigMap, as the VarnishService will take ownership of that ConfigMap after creation. Deleting any of the dependent resources will not do anything, in the same way that deleting the pod of a deployment will not. The operator will "fix" the deletion by creating a new resource to replace that which was deleted.

### Checking Status of a VarnishService Resource

The VarnishService keeps track of its current status as events occur in the system. This can be seen through the `Status` field, visible from `kubectl describe vs <your-varnishservice>`.

## Keeping Varnish Stable

Kubernetes is built on the premise that its runnable environments are ephemeral, meaning they can be created or deleted at will, with little to no effect on the overall system. In the case of Varnish, which is purely an in-memory caching layer, deleting and creating instances all the time would cause the cache to perform very poorly. Thus, there is a need to keep Varnish stable, ie tell Kubernetes that these particular runnable environments should _not_ be treated as ephemeral.

Kubernetes does not provide this functionality out of the box, but you can trick it into approximating this behavior, and that is through the concepts of guaranteed resources and affinities.

### Guaranteed Resources

The way that Kubernetes manages deployed pods on nodes is through monitoring the resources that a pod is using. Specifically, it uses the `limits` and `requests` values for `cpu` and `memory` to determine how much resources to give a pod, and when it might be OK to reschedule a pod somewhere else (namely, if a node is running out of resources and some pods are using more resources than requested). For a detailed breakdown of what `limits` and `requests` mean, [see the Kubernetes documentation on QoS](https://kubernetes.io/docs/tasks/configure-pod-container/quality-service-pod/). In QoS parlance, you want the Varnish nodes to be a "Guaranteed" QoS. In short, you want to always set the `limits` and `requests` fields, and you want `limits` and `requests` to be identical.

### Affinities

[Kubernetes allows decent control](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#inter-pod-affinity-and-anti-affinity-beta-feature) on where pods get deployed based on labels associated with pods and/or nodes. For instance, you can configure pods of the same deployment to repel each other, meaning new pods entering the deployment will try to avoid nodes that already have a pod of that type. That way, you if any one node goes down, it will only take a single pod with it. Likewise, you can configure pods to be attracted to each other, for colocation that could decrease latency between pods. Note that reading through the above linked documentation is valuable, as it goes into limitations to affinities, as well as deeply explains how they work and when to use them.

For the purposes of this Varnish deployment, you will most likely want to configure a [pod anti-affinity](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#never-co-located-in-the-same-node) (see the deployment yaml right above this section for the example) so that each pod of the varnish deployment is on a different node. Since Varnish nodes do not need to talk to each other (at least in the free versions supported by this operator), there is no need for colocation, and so you should focus on minimizing the impact of lost nodes. An example of what that might look like is in the [example annotated yaml file](/config/samples/icm_v1alpha1_varnishservice.yaml) under `spec.deployment.affinity`.

### Further Investigations

#### Taints/Tolerations

Kubernetes has a mechanism to repel pods away from a node (taints) unless the pods are specifically allowed on that node (tolerations). I am still evaluating if this could be useful in keeping Varnish stable, since it is conceivably possible that Kubernetes will sometimes just move pods around nodes to better fit things, especially in a node auto-scaling environment, even with a guaranteed resource configuration. It is unclear how often that might happen, so some testing will need to be done before further exploring taints/tolerations. At any rate, it is possible to add a toleration to the Varnish pod, in case it is needed.

### Running Varnish pods on separate IKS worker pools

This example shows how to create an IKS worker pool and make Varnish pods run strictly on its workers, one per node.

Resources:
 * [How to create IKS clusters and worker pools.](https://console.bluemix.net/docs/containers/cs_clusters.html#clusters)
 * [Taints and Tolerations](https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/)
 * [Affinity and anti-affinity](https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity)
 
1. Create a worker pool in your cluster assuming you already have a cluster called `test-cluster`

    ```bash
    $ #Find out the available zones to your cluster
    $ ibmcloud ks cluster-get --cluster test-cluster | grep "Worker Zones" # Get the 
    Worker Zones:           dal10
    $ #Find out what machine type are available in your zone  
    $ ibmcloud ks machine-types --zone dal10
    OK
    Name                      Cores   Memory   Network Speed   OS             Server Type   Storage   Secondary Storage   Trustable   
    u2c.2x4                   2       4GB      1000Mbps        UBUNTU_16_64   virtual       25GB      100GB               false   
    ms2c.4x32.1.9tb.ssd       4       32GB     10000Mbps       UBUNTU_16_64   physical      2000GB    960GB               false   
    ms2c.16x64.1.9tb.ssd      16      64GB     10000Mbps       UBUNTU_16_64   physical      2000GB    960GB               true   
    ms2c.28x256.3.8tb.ssd     28      256GB    10000Mbps       UBUNTU_16_64   physical      2000GB    1920GB              true   
       ...
    $ #Create a worker pool. 
    $ ibmcloud ks worker-pool-create --name varnish-worker-pool --cluster test-cluster --machine-type u2c.2x4 --size-per-zone 2 --hardware shared
    OK 
    $ #Verify your worker pool is created
    $ ibmcloud ks worker-pools --cluster test-cluster
    Name                  ID                                         Machine Type          Workers   
    default               91ed9433e7bf4dc7b8348ae1022f9f27-89d7d12   b2c.16x64.encrypted   3   
    varnish-worker-pool   91ed9433e7bf4dc7b8348ae1022f9f27-c5b13da   u2c.2x4.encrypted     2   
    OK
    $ #Add your zone to your worker pool. First, find out your VLAN IDs
    $ ibmcloud ks vlans --zone dal10
    OK
    ID        Name   Number   Type      Router         Supports Virtual Workers   
    2315193          1690     private   bcr02a.dal10   true   
    2315191          1425     public    fcr02a.dal10   true
    $ #Use the VLAN IDs above to add your zone to the worker pool
    $ ibmcloud ks zone-add --zone dal10 --cluster test-cluster --worker-pools varnish-worker-pool --private-vlan 2315193 --public-vlan 2315191
    OK
    $ #Verify that worker nodes provision in the zone that you've added
    $ ibmcloud ks workers --cluster test-cluster --worker-pool varnish-worker-pool
    OK
    ID                                                  Public IP   Private IP   Machine Type        State               Status                          Zone    Version   
    kube-dal10-cr91ed9433e7bf4dc7b8348ae1022f9f27-w58   -           -            u2c.2x4.encrypted   provision_pending   Preparing to provision worker   dal10   1.11.7_1543   
    kube-dal10-cr91ed9433e7bf4dc7b8348ae1022f9f27-w59   -           -            u2c.2x4.encrypted   provision_pending   -                               dal10   1.11.7_1543   
    ```
    
    Wait until your worker pool nodes change their state to `normal` and status to `Ready`.
    
    ```bash
    $ ibmcloud ks workers --cluster test-cluster --worker-pool varnish-worker-pool
    OK
    ID                                                  Public IP       Private IP      Machine Type        State    Status   Zone    Version   
    kube-dal10-cr91ed9433e7bf4dc7b8348ae1022f9f27-w58   169.61.218.68   10.94.177.179   u2c.2x4.encrypted   normal   Ready    dal10   1.11.7_1543   
    kube-dal10-cr91ed9433e7bf4dc7b8348ae1022f9f27-w59   169.61.218.94   10.94.177.180   u2c.2x4.encrypted   normal   Ready    dal10   1.11.7_1543
    ```
    
2. Taint created nodes to repel pods that don't have required toleration. 

    ```bash
    $ #Setup kubectl
    $ ibmcloud ks cluster-config --cluster test-cluster 
    OK
    The configuration for test-cluster was downloaded successfully.
    
    Export environment variables to start using Kubernetes.
    
    export KUBECONFIG=/home/me/.bluemix/plugins/container-service/clusters/test-cluster/kube-config-dal10-test-cluster.yml
    
    $ export KUBECONFIG=/home/me/.bluemix/plugins/container-service/clusters/test-cluster/kube-config-dal10-test-cluster.yml
    $ #Find your nodes using kubectl. First get your worker pool ID and then use it to select your nodes
    $ ibmcloud ks worker-pools --cluster test-cluster 
    Name                  ID                                         Machine Type          Workers   
    default               91ed9433e7bf4dc7b8348ae1022f9f27-89d7d12   b2c.16x64.encrypted   3   
    varnish-worker-pool   91ed9433e7bf4dc7b8348ae1022f9f27-c5b13da   u2c.2x4.encrypted     2   
    $ kubectl get nodes --selector ibm-cloud.kubernetes.io/worker-pool-id=91ed9433e7bf4dc7b8348ae1022f9f27-c5b13da
    NAME            STATUS   ROLES    AGE   VERSION
    10.94.177.179   Ready    <none>   16m   v1.11.7+IKS
    10.94.177.180   Ready    <none>   15m   v1.11.7+IKS
    $ #Taint those nodes
    $ kubectl taint node 10.94.177.179 role=varnish:NoSchedule #Do not schedule here not Varnish pods
    node/10.94.177.179 tainted
    $ kubectl taint node 10.94.177.179 role=varnish:NoExecute #Evict not Varnish pods if they already running here
    node/10.94.177.179 tainted
    $ kubectl taint node 10.94.177.180 role=varnish:NoSchedule #Do not schedule here not Varnish pods
    node/10.94.177.180 tainted
    $ kubectl taint node 10.94.177.180 role=varnish:NoExecute #Evict not Varnish pods if they already running here
    node/10.94.177.180 tainted
    ```
    
    This prevents all pods from scheduling on that node unless you already have pods with matching toleration
    
3. Label the nodes for the ability to schedule your varnish pods only on that nodes. Those labels will be used in your VarnishService configuration later.

    ```bash
    $ kubectl label node 10.94.177.179 role=varnish-cache
    node/10.94.177.179 labeled
    $ kubectl label node 10.94.177.180 role=varnish-cache
    node/10.94.177.180 labeled 
    ```
4. Define your VarnishService spec with necessary affinity and toleration configuration

    4.1 Define pods anti-affinity to not co-locate replicas on a node.
    
    ```yaml
    metadata:
      labels:
        role: varnish-cache
    spec:
      deployment:
        affinity:
          podAntiAffinity:
            requiredDuringSchedulingIgnoredDuringExecution:
              - labelSelector:
                  matchExpressions:
                    - key: role
                      operator: In
                      values:
                        - varnish-cache
                topologyKey: "kubernetes.io/hostname"
    ```
    That will make sure that two varnish pods doesn't get scheduled on one node. Kubernetes makes the decision based on labels we've set in the spec
    
    4.2 Define pods node affinity
    
    ```yaml
    spec:
      deployment:
        affinity:
          nodeAffinity:
            requiredDuringSchedulingIgnoredDuringExecution:
              nodeSelectorTerms:
              - matchExpressions:
                - key: role
                  operator: In
                  values:
                    - varnish-cache
    ```
    That will make kubernetes schedule varnish pods only on our worker pool nodes. The labels used here are the ones we've assigned to the node in step 3
    
    4.3 Define pods tolerations
    
    ```yaml
    spec:
      deployment:
        tolerations:
          - key: "role"
            operator: "Equal"
            value: "varnish"
            effect: "NoSchedule"
          - key: "role"
            operator: "Equal"
            value: "varnish"
            effect: "NoExecute"
    ```
    In step 2 we made our node repel all pods that don't have specific tolerations. Here we added those tolerations to be eligible for scheduling on those nodes. The values are the ones we used when tainted our nodes in step 2. 
    
5. Apply your configuration.

    This step assumes you have varnish operator [installed](#installation) and the namespace has the necessary secret [installed](#configuring-access).
    
    Complete VarnishService configuration example:
    
    ```yaml
    apiVersion: icm.ibm.com/v1alpha1
    kind: VarnishService
    metadata:
      labels:
        role: varnish-cache
      name: varnish-in-worker-pool
      namespace: varnish-ns
    spec:
      vclConfigMap:
        name: varnish-worker-pool-files
        backendsFile: backends.vcl
        defaultFile: default.vcl
      deployment:
        replicas: 2
        container:
          resources:
            limits:
              cpu: 100m
              memory: 256Mi
            requests:
              cpu: 100m
              memory: 256Mi
          readinessProbe:
            exec:
              command: [/usr/bin/varnishadm, ping]
          imagePullSecret: docker-reg-secret
        affinity:
          podAntiAffinity:
            requiredDuringSchedulingIgnoredDuringExecution:
              - labelSelector:
                  matchExpressions:
                    - key: role
                      operator: In
                      values:
                        - varnish-cache
                topologyKey: "kubernetes.io/hostname"
          nodeAffinity:
            requiredDuringSchedulingIgnoredDuringExecution:
              nodeSelectorTerms:
              - matchExpressions:
                - key: role
                  operator: In
                  values:
                    - varnish-cache
        tolerations:
          - key: "role"
            operator: "Equal"
            value: "varnish-cache"
            effect: "NoSchedule"
          - key: "role"
            operator: "Equal"
            value: "varnish-cache"
            effect: "NoExecute"
      service:
        selector:
          app: HttPerf
        varnishPort:
          name: varnish
          port: 2035
          targetPort: 8080
        varnishExporterPort:
          name: varnishexporter
          port: 9131
    ```
    Apply your configuration:
    ```bash
    $ kubectl apply -f varnish-in-worker-pool.yaml
    varnishservice.icm.ibm.com/varnish-in-worker-pool created
    ```
    Here the operator will create all pods with specified configuration
6. See your pods being scheduled strictly on your worker pool and spread across different nodes.
    ```bash
    $ kubectl get pods -n varnish-ns -o wide --selector role=varnish-cache
    NAME                                                         READY   STATUS    RESTARTS   AGE   IP               NODE            NOMINATED NODE
    varnish-in-worker-pool-varnish-deployment-78c9b6f5bf-kqg72   1/1     Running   0          6m    172.30.244.65    10.94.177.179   <none>
    varnish-in-worker-pool-varnish-deployment-78c9b6f5bf-pqtzv   1/1     Running   0          6m    172.30.136.129   10.94.177.180   <none>

    ```
    Check the `NODE` column. The value will be different for each pod.
    
    Note that you won't be able to run more pods than you have nodes. The anti-affinity rule will not allow two pods being co-located on one node.
    This behaviour can be changed by using an anti-affinity type called `preferredDuringSchedulingIgnoredDuringExecution`: 
    
    ```yaml
    spec:
      deployment:
        affinity:
          podAntiAffinity:
            preferredDuringSchedulingIgnoredDuringExecution:
            - weight: 1
              podAffinityTerm:
                labelSelector:
                  matchExpressions:
                  - key: role
                    operator: In
                    values:
                    - varnish-cache
                topologyKey: "kubernetes.io/hostname"
    ```
     It will still ask Kubernetes to spread pods onto different nodes but also allow to co-locate them if there are more pods than nodes.
     
    Also keep in mind that in such configuration the pods can be scheduled to your worker pool only. If the worker pool is deleted the pods will hang in `Pending` state until new nodes with the same configuration are added to the cluster.