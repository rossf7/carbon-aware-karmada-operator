package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Reasons why operator is in degraded status
const (
	ReasonSucceeded                       = "OperatorSucceeded"
	ReasonStatusUpdateFailed              = "OperatorStatusUpdateFailed"
	ReasonTargetUpdateFailed              = "OperatorTargetUpdateFailed"
	ReasonTargetNotFound                  = "OperatorTargetNotFound"
	ReasonTargetFetchError                = "OperatorTargetFetchError"
	ReasonCarbonIntensityFetchError       = "OperatorCarbonIntensityFetchError"
	ReasonCarbonIntensityLocationNotFound = "OperatorCarbonIntensityLocationNotFound"
)

// CarbonAwareKarmadaPolicySpec defines the desired state of CarbonAwareKarmadaPolicy
type CarbonAwareKarmadaPolicySpec struct {
	// ClusterLocations is an array of member clusters and their locations using
	// the location codes supported by the carbon intensity API being used.
	// +kubebuilder:validation:Required
	ClusterLocations []ClusterLocation `json:"clusterLocations"`

	// DesiredClusters is the number of member clusters to select.
	// +kubebuilder:validation:Required
	DesiredClusters *int32 `json:"desiredClusters"`

	// KarmadaTarget is the type of the karmada object to update.
	// +kubebuilder:validation:Required
	KarmadaTarget KarmadaTarget `json:"karmadaTarget"`

	// KarmadaTargetRef is the reference to the karmada object to update.
	// +kubebuilder:validation:Required
	KarmadaTargetRef KarmadaTargetRef `json:"karmadaTargetRef"`
}

// CarbonAwareKarmadaPolicyStatus defines the observed state of CarbonAwareKarmadaPolicy
type CarbonAwareKarmadaPolicyStatus struct {
	// ActiveClusters is an array of member cluster names.
	ActiveClusters []string `json:"activeClusters"`

	// Clusters is an array of member cluster statuses including location and carbon intensity.
	Clusters []ClusterStatus `json:"clusters"`

	// Provider of carbon intensity data.
	Provider string `json:"provider"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CarbonAwareKarmadaPolicy is the Schema for the carbonawarekarmadapolicies API
type CarbonAwareKarmadaPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CarbonAwareKarmadaPolicySpec   `json:"spec,omitempty"`
	Status CarbonAwareKarmadaPolicyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CarbonAwareKarmadaPolicyList contains a list of CarbonAwareKarmadaPolicy
type CarbonAwareKarmadaPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CarbonAwareKarmadaPolicy `json:"items"`
}

// ClusterLocation represents a member cluster and its physical location
// so the carbon intensity for this location can be retrieved.
type ClusterLocation struct {
	// location of the karmada member cluster
	// +kubebuilder:validation:Required
	Location string `json:"location"`

	// name of the karmada member cluster
	// +kubebuilder:validation:Required
	Name string `json:"name"`
}

// KarmadaTarget represents the type of the Karmada policy
// Only one of the following Karmada policies is supported:
// - clusterpropagationpolicies.policy.karmada.io
// - propagationpolicies.policy.karmada.io
// +kubebuilder:validation:Enum=clusterpropagationpolicies.policy.karmada.io;propagationpolicies.policy.karmada.io
type KarmadaTarget string

const (
	ClusterPropagationPolicy KarmadaTarget = "clusterpropagationpolicies.policy.karmada.io"
	PropagationPolicy        KarmadaTarget = "propagationpolicies.policy.karmada.io"
)

// KarmadaTargetRef represents the Karmada policy to update
type KarmadaTargetRef struct {
	// name of the karmada policy
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// namespace of the karmada policy
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`
}

type CarbonIntensity struct {
	Units     string `json:"units"`
	ValidFrom string `json:"validFrom"`
	ValidTo   string `json:"validTo"`
	Value     string `json:"value"`
}

// ClusterStatus represents a member cluster and its physical location
// so the carbon intensity for this location can be retrieved.
type ClusterStatus struct {
	// carbon intensity for this location
	CarbonIntensity CarbonIntensity `json:"carbonIntensity"`

	// location of the karmada member cluster
	// +kubebuilder:validation:Required
	Location string `json:"location"`

	// name of the karmada member cluster
	// +kubebuilder:validation:Required
	Name string `json:"name"`
}

func init() {
	SchemeBuilder.Register(&CarbonAwareKarmadaPolicy{}, &CarbonAwareKarmadaPolicyList{})
}
