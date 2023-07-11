//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CarbonAwareKarmadaPolicy) DeepCopyInto(out *CarbonAwareKarmadaPolicy) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CarbonAwareKarmadaPolicy.
func (in *CarbonAwareKarmadaPolicy) DeepCopy() *CarbonAwareKarmadaPolicy {
	if in == nil {
		return nil
	}
	out := new(CarbonAwareKarmadaPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CarbonAwareKarmadaPolicy) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CarbonAwareKarmadaPolicyList) DeepCopyInto(out *CarbonAwareKarmadaPolicyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CarbonAwareKarmadaPolicy, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CarbonAwareKarmadaPolicyList.
func (in *CarbonAwareKarmadaPolicyList) DeepCopy() *CarbonAwareKarmadaPolicyList {
	if in == nil {
		return nil
	}
	out := new(CarbonAwareKarmadaPolicyList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CarbonAwareKarmadaPolicyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CarbonAwareKarmadaPolicySpec) DeepCopyInto(out *CarbonAwareKarmadaPolicySpec) {
	*out = *in
	if in.ActiveClusters != nil {
		in, out := &in.ActiveClusters, &out.ActiveClusters
		*out = new(int32)
		**out = **in
	}
	if in.ClusterLocations != nil {
		in, out := &in.ClusterLocations, &out.ClusterLocations
		*out = make([]ClusterLocation, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	out.KarmadaPolicyRef = in.KarmadaPolicyRef
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CarbonAwareKarmadaPolicySpec.
func (in *CarbonAwareKarmadaPolicySpec) DeepCopy() *CarbonAwareKarmadaPolicySpec {
	if in == nil {
		return nil
	}
	out := new(CarbonAwareKarmadaPolicySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CarbonAwareKarmadaPolicyStatus) DeepCopyInto(out *CarbonAwareKarmadaPolicyStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CarbonAwareKarmadaPolicyStatus.
func (in *CarbonAwareKarmadaPolicyStatus) DeepCopy() *CarbonAwareKarmadaPolicyStatus {
	if in == nil {
		return nil
	}
	out := new(CarbonAwareKarmadaPolicyStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClusterLocation) DeepCopyInto(out *ClusterLocation) {
	*out = *in
	if in.GeoLocation != nil {
		in, out := &in.GeoLocation, &out.GeoLocation
		*out = make([]resource.Quantity, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClusterLocation.
func (in *ClusterLocation) DeepCopy() *ClusterLocation {
	if in == nil {
		return nil
	}
	out := new(ClusterLocation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KarmadaPolicyRef) DeepCopyInto(out *KarmadaPolicyRef) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KarmadaPolicyRef.
func (in *KarmadaPolicyRef) DeepCopy() *KarmadaPolicyRef {
	if in == nil {
		return nil
	}
	out := new(KarmadaPolicyRef)
	in.DeepCopyInto(out)
	return out
}
