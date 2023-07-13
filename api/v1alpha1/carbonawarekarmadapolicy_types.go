package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CarbonAwareKarmadaPolicySpec defines the desired state of CarbonAwareKarmadaPolicy
type CarbonAwareKarmadaPolicySpec struct {
	// array of member clusters and their physical locations
	// +kubebuilder:validation:Required
	ClusterLocations []ClusterLocation `json:"clusterLocations"`

	// number of member clusters to propagate resources to.
	// +kubebuilder:validation:Required
	DesiredClusters *int32 `json:"desiredClusters"`

	// type of the karmada object to scale
	// +kubebuilder:validation:Required
	KarmadaTarget KarmadaTarget `json:"karmadaTarget"`

	// reference to the karmada object to scale
	// +kubebuilder:validation:Required
	KarmadaTargetRef KarmadaTargetRef `json:"karmadaTargetRef"`
}

// CarbonAwareKarmadaPolicyStatus defines the observed state of CarbonAwareKarmadaPolicy
type CarbonAwareKarmadaPolicyStatus struct {
	ActiveClusters []string        `json:"activeClusters"`
	Clusters       []ClusterStatus `json:"clusters"`
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

type ClusterCarbonIntensityStatus struct {
	Units     string `json:"units"`
	ValidFrom string `json:"validFrom"`
	ValidTo   string `json:"validTo"`
	Value     string `json:"value"`
}

type ClusterStatus struct {
	CarbonIntensity ClusterCarbonIntensityStatus `json:"carbonIntensity"`
	Location        string                       `json:"location"`
	Name            string                       `json:"name"`
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

// KarmadaTargetRef represents the Karmada object to scale
type KarmadaTargetRef struct {
	// name of the karmada policy
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// namespace of the karmada policy
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`
}

func init() {
	SchemeBuilder.Register(&CarbonAwareKarmadaPolicy{}, &CarbonAwareKarmadaPolicyList{})
}
