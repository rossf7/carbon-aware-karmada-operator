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
	requeueInterval time.Duration = 5 * time.Minute
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
//+kubebuilder:rbac:groups=policy.karmada.io,resources=clusterpropagationpolicies,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=policy.karmada.io,resources=propagationpolicies,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *CarbonAwareKarmadaPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	karmadav1alpha1.AddToScheme(r.Scheme)

	carbonAwareKarmadaPolicy := &carbonawarev1alpha1.CarbonAwareKarmadaPolicy{}
	err := r.Get(ctx, req.NamespacedName, carbonAwareKarmadaPolicy)
	if err != nil && apierrors.IsNotFound(err) {
		logger.Error(err, "unable to find carbon aware karmada policy")
		return ctrl.Result{RequeueAfter: requeueInterval}, client.IgnoreNotFound(err)
	} else if err != nil {
		logger.Error(err, "failed to find carbon aware karmada policy")
		ReconcileErrorsTotal.WithLabelValues(carbonAwareKarmadaPolicy.Name).Inc()
		return ctrl.Result{RequeueAfter: requeueInterval}, err
	}

	ReconcilesTotal.WithLabelValues(carbonAwareKarmadaPolicy.Name).Inc()

	clusters := []ClusterCarbonIntensity{}

	for _, loc := range carbonAwareKarmadaPolicy.Spec.ClusterLocations {
		clusterCarbonIntensity, err := r.CarbonIntensityFetcher.Fetch(ctx, loc.Name, loc.Location)
		if err != nil {
			logger.Error(err, "unable to get carbon intensity", "location", loc.Location)
			ReconcileErrorsTotal.WithLabelValues(carbonAwareKarmadaPolicy.Name).Inc()
			return ctrl.Result{RequeueAfter: requeueInterval}, err
		}
		clusters = append(clusters, clusterCarbonIntensity)
	}

	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].CarbonIntensity.Value < clusters[j].CarbonIntensity.Value
	})

	activeClusters := []string{}
	clusterStatuses := []carbonawarev1alpha1.ClusterStatus{}
	desiredClusters := int(*carbonAwareKarmadaPolicy.Spec.DesiredClusters)

	for i, c := range clusters {
		var active bool

		if i < desiredClusters {
			activeClusters = append(activeClusters, c.ClusterName)
			active = true
		}

		status := carbonawarev1alpha1.ClusterStatus{
			CarbonIntensity: carbonawarev1alpha1.ClusterCarbonIntensityStatus{
				Units:     c.CarbonIntensity.Units,
				ValidFrom: c.CarbonIntensity.ValidFrom.Format(time.RFC3339),
				ValidTo:   c.CarbonIntensity.ValidTo.Format(time.RFC3339),
				Value:     fmt.Sprintf("%.2f", c.CarbonIntensity.Value),
			},
			Location: c.CarbonIntensity.Location,
			Name:     c.ClusterName,
		}
		clusterStatuses = append(clusterStatuses, status)
		CarbonIntensityMetric.WithLabelValues(c.ClusterName,
			c.CarbonIntensity.Location,
			strconv.FormatBool(active)).Set(c.CarbonIntensity.Value)
	}

	switch {
	case strings.Contains(string(carbonAwareKarmadaPolicy.Spec.KarmadaTarget), "clusterpropagationpolicies"):
		clusterPropagationPolicy := &karmadav1alpha1.ClusterPropagationPolicy{}
		err = r.Get(ctx, types.NamespacedName{Name: carbonAwareKarmadaPolicy.Spec.KarmadaTargetRef.Name}, clusterPropagationPolicy)
		if err != nil && apierrors.IsNotFound(err) {
			logger.Error(err, "unable to find cluster propagation policy")
			return ctrl.Result{RequeueAfter: requeueInterval}, err
		} else if err != nil {
			logger.Error(err, "failed to find cluster propagation policy")
			ReconcileErrorsTotal.WithLabelValues(carbonAwareKarmadaPolicy.Name).Inc()
			return ctrl.Result{RequeueAfter: requeueInterval}, err
		}

		if clusterPropagationPolicy.Spec.Placement.ClusterAffinity == nil {
			clusterPropagationPolicy.Spec.Placement.ClusterAffinity = &karmadav1alpha1.ClusterAffinity{
				ClusterNames: activeClusters,
			}
		} else {
			clusterPropagationPolicy.Spec.Placement.ClusterAffinity.ClusterNames = activeClusters
		}
		err = r.Update(ctx, clusterPropagationPolicy)
		if err != nil {
			logger.Error(err, "unable to update cluster propagation policy")
			ReconcileErrorsTotal.WithLabelValues(carbonAwareKarmadaPolicy.Name).Inc()
			return ctrl.Result{RequeueAfter: requeueInterval}, err
		}
	case strings.Contains(string(carbonAwareKarmadaPolicy.Spec.KarmadaTarget), "propagationpolicies"):
		propagationPolicy := &karmadav1alpha1.PropagationPolicy{}
		err = r.Get(ctx, types.NamespacedName{Name: carbonAwareKarmadaPolicy.Spec.KarmadaTargetRef.Name,
			Namespace: carbonAwareKarmadaPolicy.Spec.KarmadaTargetRef.Namespace}, propagationPolicy)
		if err != nil && apierrors.IsNotFound(err) {
			logger.Error(err, "unable to find propagation policy")
			return ctrl.Result{RequeueAfter: requeueInterval}, err
		} else if err != nil {
			logger.Error(err, "failed to find propagation policy")
			ReconcileErrorsTotal.WithLabelValues(carbonAwareKarmadaPolicy.Name).Inc()
			return ctrl.Result{RequeueAfter: requeueInterval}, err
		}

		if propagationPolicy.Spec.Placement.ClusterAffinity == nil {
			propagationPolicy.Spec.Placement.ClusterAffinity = &karmadav1alpha1.ClusterAffinity{
				ClusterNames: activeClusters,
			}
		} else {
			propagationPolicy.Spec.Placement.ClusterAffinity.ClusterNames = activeClusters
		}
		err = r.Update(ctx, propagationPolicy)
		if err != nil {
			logger.Error(err, "unable to update propagation policy")
			ReconcileErrorsTotal.WithLabelValues(carbonAwareKarmadaPolicy.Name).Inc()
			return ctrl.Result{RequeueAfter: requeueInterval}, err
		}
	}

	carbonAwareKarmadaPolicy.Status.ActiveClusters = activeClusters
	carbonAwareKarmadaPolicy.Status.Clusters = clusterStatuses
	err = r.Status().Update(ctx, carbonAwareKarmadaPolicy)
	if err != nil {
		logger.Error(err, "unable to update carbon aware policy status")
		ReconcileErrorsTotal.WithLabelValues(carbonAwareKarmadaPolicy.Name).Inc()
		return ctrl.Result{RequeueAfter: requeueInterval}, err
	}

	return ctrl.Result{RequeueAfter: requeueInterval}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CarbonAwareKarmadaPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&carbonawarev1alpha1.CarbonAwareKarmadaPolicy{}).
		Complete(r)
}
