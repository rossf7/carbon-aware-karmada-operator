# carbon-aware-karmada-operator

This is a Kubernetes operator that does carbon aware spatial shifting using
[Karmada](https://github.com/karmada-io/karmada) so workloads run in clusters
at physical locations with the lowest carbon intensity.

Carbon intensity measures how much carbon (CO2 equivalent) is emitted per kilowatt-hour (KWh) of electricity consumed. Carbon intensity varies by location depending on the electricity grid and by time depending on how much renewable energy is available. The operator uses the [grid-intensity-go](https://github.com/thegreenwebfoundation/grid-intensity-go) library from [The Green Web Foundation](https://www.thegreenwebfoundation.org/) to fetch carbon intensity data from the [Electricity Maps](https://www.electricitymaps.com/) or [WattTime](https://www.watttime.org/) APIs.

:warning: This operator is experimental and not intended for production usage yet.

## How Does It Work?

[Karmada](https://karmada.io/) runs in a control plane cluster and can schedule workloads across multiple member clusters. 

Karmada allows you to define a propagation policy that defines which resources to schedule in the member
clusters. In this case it's an nginx deployment but it could be any kubernetes resource.

```yaml
apiVersion: policy.karmada.io/v1alpha1
kind: PropagationPolicy
metadata:
  name: nginx-propagation
spec:
  resourceSelectors:
    - apiVersion: apps/v1
      kind: Deployment
      name: nginx
  placement:
    replicaScheduling:
      replicaSchedulingType: Divided
```

The `carbon-aware-karmada-operator` extends Karmada by letting you define a carbon aware policy.

```yaml
apiVersion: carbonaware.rossf7.github.io/v1alpha1
kind: CarbonAwareKarmadaPolicy
metadata:
  name: nginx-policy
spec:
  clusterLocations:
  - name: prd-de-01
    location: DE
  - name: prd-fr-01
    location: FR
  desiredClusters: 1
  karmadaTarget: propagationpolicies.policy.karmada.io
  karmadaTargetRef:
    name: nginx-propagation
    namespace: default
```

- `.spec.clusterLocations` is an array of member clusters and their locations using the location
codes supported by the carbon intensity API being used.
- `.spec.desiredClusters` is how many member clusters to select. Clusters are ranked based on their
current carbon intensity.
- `.spec.karmadaTarget` and `.spec.karmadaTargetRef` is the Karmada `PropagationPolicy` or
`ClusterPropagationPolicy` to update.

The `carbon-aware-karmada-operator` sets the cluster affinity in the propagation policy. Karmada then
schedules the resources in the selected member clusters.

```yaml
apiVersion: policy.karmada.io/v1alpha1
kind: PropagationPolicy
metadata:
  name: nginx-propagation
spec:
  placement:
    clusterAffinity:
      clusterNames:
      - prd-fr-01
```

## Quick Start

1. Follow the Karmada [quick start](https://github.com/karmada-io/karmada#install-the-karmada-control-plane)
to create 4 [kind](https://sigs.k8s.io/kind) clusters. A control plane cluster and 3 member clusters.

2. Configure your kubectl to connect to the control plane cluster and the Karmada API server.

```sh
export KUBECONFIG=~/.kube/karmada.config
kubectl config use-context karmada-apiserver
```

3. The default carbon intensity provider is [ElectricityMaps](https://api-portal.electricitymaps.com/)
who have a free tier for non commercial use.

Register for an API key and set the environment variables. Use the `/zones` endpoint
https://api.electricitymap.org/v3/zones to see the supported locations.

```sh
export ELECTRICITY_MAP_API_TOKEN=******
export ELECTRICITY_MAP_API_URL=https://api-access.electricitymaps.com/free-tier/
```

See [Providers](#providers) for more config options.

4. Clone this repo.

```sh
git clone https://github.com/rossf7/carbon-aware-karmada-operator.git
cd carbon-aware-karmada-operator
```

5. Install the `CarbonAwareKarmadaPolicy` CRD.

```sh
make install
```

6. Start the controller.

```sh
make run
```

7. Create the sample resources.

```sh
kubectl apply -f samples/nginx/
```

8. Get the custom resources to see which clusters were selected.

```sh
kubectl get carbonawarekarmadapolicies carbon-aware-nginx-policy -o yaml
kubectl get propagationpolicies.policy.karmada.io nginx-propagation -o yaml
```

9. Finally check the nginx deployment is scheduled in one of selected member clusters.

```sh
export KUBECONFIG=~/.kube/members.config
kubectl config use-context member1
kubectl get deploy nginx
```

## Providers

The following providers are supported.

### Electricity Maps

When using the [API portal](https://api-portal.electricitymaps.com) (api-portal.electricitymaps.com)
you'll need to set both the API token and API URL.

When using the free tier the URL is `https://api-access.electricitymaps.com/free-tier/`.
For paid plans the URL will be displayed in the API portal.

```sh
export ELECTRICITY_MAP_API_TOKEN=******
export ELECTRICITY_MAP_API_URL=******
```

Use the `/zones` endpoint https://api.electricitymap.org/v3/zones to see the supported locations.

### WattTime

[Register](https://www.watttime.org/api-documentation/#authentication) for an API account and
set the env var and provider name.

```sh
export WATT_TIME_API_USER=******
export WATT_TIME_API_PASSWORD=******
go run cmd/main.go -provider-name WattTime
```

Use the `/ba-from-loc` [endpoint](https://www.watttime.org/api-documentation/#determine-grid-region)
to see the supported locations.

## Credit

- https://learn.greensoftware.foundation/carbon-awareness/
- https://github.com/Azure/carbon-aware-keda-operator
- https://github.com/karmada-io/karmada
- https://github.com/thegreenwebfoundation/grid-intensity-go

## License

carbon-aware-karmada-operator is under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.
