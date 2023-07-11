/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"os"
	"sort"

	karmadav1alpha1 "github.com/karmada-io/karmada/pkg/apis/policy/v1alpha1"
	"github.com/thegreenwebfoundation/grid-intensity-go/pkg/provider"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	carbonawarev1alpha1 "github.com/rossf7/carbon-aware-karmada-operator/api/v1alpha1"
)

// CarbonAwareKarmadaPolicyReconciler reconciles a CarbonAwareKarmadaPolicy object
type CarbonAwareKarmadaPolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=carbonaware.rossf7.github.io,resources=carbonawarekarmadapolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=carbonaware.rossf7.github.io,resources=carbonawarekarmadapolicies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=carbonaware.rossf7.github.io,resources=carbonawarekarmadapolicies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CarbonAwareKarmadaPolicy object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *CarbonAwareKarmadaPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	karmadav1alpha1.AddToScheme(r.Scheme)

	logger.Info("processing CR")

	carbonAwareKarmadaPolicy := &carbonawarev1alpha1.CarbonAwareKarmadaPolicy{}
	err := r.Get(ctx, req.NamespacedName, carbonAwareKarmadaPolicy)
	if err != nil {
		logger.Error(err, "unable to find carbonawarekarmadapolicy")
	}

	logger.Info("got custom resource", "policy", carbonAwareKarmadaPolicy)

	c := provider.WattTimeConfig{
		APIUser:     os.Getenv("WATT_TIME_API_USER"),
		APIPassword: os.Getenv("WATT_TIME_API_PASSWORD"),
	}
	p, err := provider.NewWattTime(c)
	if err != nil {
		logger.Error(err, "unable to create watt time provider")
	}

	locations := map[string]float64{}

	for _, loc := range carbonAwareKarmadaPolicy.Spec.ClusterLocations {
		carbonIntensity, err := p.GetCarbonIntensity(ctx, loc.Location)
		if err != nil {
			logger.Error(err, "unable to get carbon intensity", "location", loc.Location)
		}

		logger.Info("cluster location", "loc", loc, "carbon intensity", carbonIntensity)
		locations[loc.Name] = carbonIntensity[0].Value
	}

	type kv struct {
		Key   string
		Value float64
	}

	var clusters []kv
	for k, v := range locations {
		clusters = append(clusters, kv{k, v})
	}

	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].Value < clusters[j].Value
	})

	activeClusters := int(*carbonAwareKarmadaPolicy.Spec.ActiveClusters)
	clusterNames := []string{}

	for i := 0; i < activeClusters; i++ {
		clusterNames = append(clusterNames, clusters[i].Key)
	}

	propagationPolicy := &karmadav1alpha1.PropagationPolicy{}
	err = r.Get(ctx, types.NamespacedName{Name: carbonAwareKarmadaPolicy.Spec.KarmadaPolicyRef.Name,
		Namespace: carbonAwareKarmadaPolicy.Spec.KarmadaPolicyRef.Namespace}, propagationPolicy)
	if err != nil {
		logger.Error(err, "unable to find propagation policy")
	}

	logger.Info("got propagation policy", "policy", propagationPolicy)

	propagationPolicy.Spec.Placement.ClusterAffinity.ClusterNames = clusterNames

	err = r.Update(ctx, propagationPolicy)
	if err != nil {
		logger.Error(err, "unable to update propagation policy")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CarbonAwareKarmadaPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&carbonawarev1alpha1.CarbonAwareKarmadaPolicy{}).
		Complete(r)
}
