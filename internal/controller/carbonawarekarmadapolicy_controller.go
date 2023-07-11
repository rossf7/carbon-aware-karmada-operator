package controller

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	karmadav1alpha1 "github.com/karmada-io/karmada/pkg/apis/policy/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	carbonawarev1alpha1 "github.com/rossf7/carbon-aware-karmada-operator/api/v1alpha1"
)

const (
	requeueInterval = time.Duration(5 * time.Minute)
)

// CarbonAwareKarmadaPolicyReconciler reconciles a CarbonAwareKarmadaPolicy object
type CarbonAwareKarmadaPolicyReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	CarbonIntensityFetcher
}

//+kubebuilder:rbac:groups=carbonaware.rossf7.github.io,resources=carbonawarekarmadapolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=carbonaware.rossf7.github.io,resources=carbonawarekarmadapolicies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=carbonaware.rossf7.github.io,resources=carbonawarekarmadapolicies/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *CarbonAwareKarmadaPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	karmadav1alpha1.AddToScheme(r.Scheme)

	// Get carbon aware policy CR.
	carbonAwareKarmadaPolicy := &carbonawarev1alpha1.CarbonAwareKarmadaPolicy{}
	err := r.Get(ctx, req.NamespacedName, carbonAwareKarmadaPolicy)
	if err != nil {
		logger.Error(err, "unable to find carbonawarekarmadapolicy")
		r.Recorder.Event(carbonAwareKarmadaPolicy, "Warning", "NoCustomResource", fmt.Sprintf("Unable to find carbonawarekarmadapolicy %s", req.NamespacedName))
		return ctrl.Result{RequeueAfter: requeueInterval}, err
	}

	ReconcilesTotal.WithLabelValues(carbonAwareKarmadaPolicy.Name).Inc()

	clusters := []ClusterCarbonIntensity{}

	// Fetch carbon intensity for each cluster location.
	for _, loc := range carbonAwareKarmadaPolicy.Spec.ClusterLocations {
		cluster, err := r.CarbonIntensityFetcher.Fetch(ctx, loc.Name, loc.Location)
		if err != nil {
			logger.Error(err, "failed to fetch carbon intensity")
			return ctrl.Result{RequeueAfter: requeueInterval}, err
		}
		if cluster.CarbonIntensity.Location == loc.Location {
			clusters = append(clusters, cluster)
		} else {
			r.Recorder.Event(carbonAwareKarmadaPolicy, "Warning", carbonawarev1alpha1.ReasonCarbonIntensityLocationNotFound, fmt.Sprintf("failed to get carbon intensity for location %s", loc.Location))
		}
	}

	// Sort cluster locations by lowest carbon intensity.
	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].CarbonIntensity.Value < clusters[j].CarbonIntensity.Value
	})

	desiredClusters := int(*carbonAwareKarmadaPolicy.Spec.DesiredClusters)
	activeClusterNames := []string{}
	clusterStatuses := []carbonawarev1alpha1.ClusterStatus{}

	for i, cluster := range clusters {
		var active bool

		clusterName := cluster.ClusterName
		location := cluster.CarbonIntensity.Location
		value := cluster.CarbonIntensity.Value

		if i < desiredClusters {
			activeClusterNames = append(activeClusterNames, clusterName)
			active = true
		}
		status := carbonawarev1alpha1.ClusterStatus{
			CarbonIntensity: carbonawarev1alpha1.CarbonIntensity{
				Units:     cluster.CarbonIntensity.Units,
				ValidFrom: cluster.CarbonIntensity.ValidFrom.Format(time.RFC3339),
				ValidTo:   cluster.CarbonIntensity.ValidTo.Format(time.RFC3339),
				Value:     fmt.Sprintf("%.1f", value),
			},
			Location: location,
			Name:     clusterName,
		}
		clusterStatuses = append(clusterStatuses, status)

		CarbonIntensityMetric.WithLabelValues(clusterName, location, strconv.FormatBool(active)).Set(value)
	}

	logger.Info(fmt.Sprintf("selected %d active clusters", len(activeClusterNames)), "clusters", activeClusterNames)
	karmadaGVK := string(carbonAwareKarmadaPolicy.Spec.KarmadaTarget)

	switch {
	case strings.Contains(karmadaGVK, "clusterpropagationpolicies"):
		// Set cluster affinity for cluster propagation policy CR.
		clusterPropagationPolicy := &karmadav1alpha1.ClusterPropagationPolicy{}
		err = r.Get(ctx, types.NamespacedName{Name: carbonAwareKarmadaPolicy.Spec.KarmadaTargetRef.Name}, clusterPropagationPolicy)
		if err != nil && apierrors.IsNotFound(err) {
			logger.Error(err, "unable to get cluster propagation policy")
			return ctrl.Result{RequeueAfter: requeueInterval}, err
		} else if err != nil {
			ReconcileErrorsTotal.WithLabelValues(carbonAwareKarmadaPolicy.Name).Inc()
			logger.Error(err, "failed to get cluster propagation policy")
			return ctrl.Result{RequeueAfter: requeueInterval}, err
		}

		logger.Info("got cluster propagation policy", "policy", clusterPropagationPolicy)

		if clusterPropagationPolicy.Spec.Placement.ClusterAffinity == nil {
			clusterPropagationPolicy.Spec.Placement.ClusterAffinity = &karmadav1alpha1.ClusterAffinity{
				ClusterNames: activeClusterNames,
			}
		} else {
			clusterPropagationPolicy.Spec.Placement.ClusterAffinity.ClusterNames = activeClusterNames
		}
		err = r.Update(ctx, clusterPropagationPolicy)
		if err != nil {
			ReconcileErrorsTotal.WithLabelValues(carbonAwareKarmadaPolicy.Name).Inc()
			logger.Error(err, "unable to update cluster propagation policy")
			return ctrl.Result{RequeueAfter: requeueInterval}, err
		}
	case strings.Contains(karmadaGVK, "propagationpolicies"):
		// Set cluster affinity for namespace scoped propagation policy CR.
		propagationPolicy := &karmadav1alpha1.PropagationPolicy{}
		err = r.Get(ctx, types.NamespacedName{Name: carbonAwareKarmadaPolicy.Spec.KarmadaTargetRef.Name,
			Namespace: carbonAwareKarmadaPolicy.Spec.KarmadaTargetRef.Namespace}, propagationPolicy)
		if err != nil && apierrors.IsNotFound(err) {
			logger.Error(err, "unable to get propagation policy")
			return ctrl.Result{RequeueAfter: requeueInterval}, err
		} else if err != nil {
			ReconcileErrorsTotal.WithLabelValues(carbonAwareKarmadaPolicy.Name).Inc()
			logger.Error(err, "failed to get propagation policy")
			return ctrl.Result{RequeueAfter: requeueInterval}, err
		}

		logger.Info("got propagation policy", "policy", propagationPolicy)

		if propagationPolicy.Spec.Placement.ClusterAffinity == nil {
			propagationPolicy.Spec.Placement.ClusterAffinity = &karmadav1alpha1.ClusterAffinity{
				ClusterNames: activeClusterNames,
			}
		} else {
			propagationPolicy.Spec.Placement.ClusterAffinity.ClusterNames = activeClusterNames
		}
		err = r.Update(ctx, propagationPolicy)
		if err != nil {
			ReconcileErrorsTotal.WithLabelValues(carbonAwareKarmadaPolicy.Name).Inc()
			logger.Error(err, "unable to update propagation policy")
			return ctrl.Result{RequeueAfter: time.Duration(requeueInterval)}, err
		}
	}

	// Set status for carbon aware policy CR.
	carbonAwareKarmadaPolicy.Status.ActiveClusters = activeClusterNames
	carbonAwareKarmadaPolicy.Status.Clusters = clusterStatuses
	carbonAwareKarmadaPolicy.Status.Provider = r.CarbonIntensityFetcher.Provider()
	err = r.Status().Update(ctx, carbonAwareKarmadaPolicy)
	if err != nil {
		ReconcileErrorsTotal.WithLabelValues(carbonAwareKarmadaPolicy.Name).Inc()
		logger.Error(err, "unable to update status for carbon aware karmada policy")
		return ctrl.Result{RequeueAfter: requeueInterval}, err
	}

	// Record the successful reconcile event.
	r.Recorder.Event(carbonAwareKarmadaPolicy, "Normal", "ActiveClustersReconciled", fmt.Sprintf("Successfully set active clusters for %s to %s", carbonAwareKarmadaPolicy.Spec.KarmadaTargetRef.Name, strings.Join(activeClusterNames, ",")))

	return ctrl.Result{RequeueAfter: requeueInterval}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CarbonAwareKarmadaPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&carbonawarev1alpha1.CarbonAwareKarmadaPolicy{}).
		Complete(r)
}
