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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CarbonAwareKarmadaPolicySpec defines the desired state of CarbonAwareKarmadaPolicy
type CarbonAwareKarmadaPolicySpec struct {
	// number of member clusters to propagate resources to.
	// +kubebuilder:validation:Required
	ActiveClusters *int32 `json:"activeClusters"`

	// array of member clusters and their physical locations
	// +kubebuilder:validation:Required
	ClusterLocations []ClusterLocation `json:"clusterLocations"`

	// type of the karmada object to scale
	// +kubebuilder:validation:Required
	KarmadaPolicy KarmadaPolicy `json:"karmadaPolicy"`

	// reference to the karmada object to scale
	// +kubebuilder:validation:Required
	KarmadaPolicyRef KarmadaPolicyRef `json:"karmadaPolicyRef"`
}

// CarbonAwareKarmadaPolicyStatus defines the observed state of CarbonAwareKarmadaPolicy
type CarbonAwareKarmadaPolicyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
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
	// geo location of the karmada member cluster
	// +kubebuilder:validation:Required
	GeoLocation []resource.Quantity `json:"geoLocation"`

	// location of the karmada member cluster
	// +kubebuilder:validation:Required
	Location string `json:"location"`

	// name of the karmada member cluster
	// +kubebuilder:validation:Required
	Name string `json:"name"`
}

// KarmadaPolicy represents the type of the Karmada policy
// Only one of the following Karmada policies is supported:
// - clusterpropagationpolicies.policy.karmada.io
// - propagationpolicies.policy.karmada.io
// +kubebuilder:validation:Enum=clusterpropagationpolicies.policy.karmada.io;propagationpolicies.policy.karmada.io
type KarmadaPolicy string

// KarmadaPolicyRef represents the Karmada object to scale
type KarmadaPolicyRef struct {
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
