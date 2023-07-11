package controller

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	CarbonIntensityMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "carbon_aware_karmada_operator_carbon_intensity",
			Help: "Carbon intensity",
		},
		[]string{"cluster", "location", "active"},
	)

	ReconcilesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "carbon_aware_karmada_operator_reconciles_total",
			Help: "Total number of reconciles",
		},
		[]string{"app"},
	)

	ReconcileErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "carbon_aware_karmada_operator_reconcile_errors_total",
			Help: "Total number of reconcile errors",
		},
		[]string{"app"},
	)
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(CarbonIntensityMetric)
	metrics.Registry.MustRegister(ReconcilesTotal)
	metrics.Registry.MustRegister(ReconcileErrorsTotal)
}
