/*
Copyright 2018 The Knative Authors.

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
	"encoding/json"
	"sort"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "github.com/knative/pkg/apis"
	autoscalev1 "k8s.io/api/autoscaling/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// TODO
type PodAutoscaler struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PodAutoscalerSpec   `json:"spec"`
	Status PodAutoscalerStatus `json:"status"`
}

// TODO(mattmoor): Check that PodAutoscaler may be validated and defaulted.
// var _ apis.Validatable = (*PodAutoscaler)(nil)
// var _ apis.Defaultable = (*PodAutoscaler)(nil)

// TODO(mattmoor): PodAutoscalerSpec represents ...
type PodAutoscalerSpec struct {
	// TODO: Generation does not work correctly with CRD. They are scrubbed
	// by the APIserver (https://github.com/kubernetes/kubernetes/issues/58778)
	// So, we add Generation here. Once that gets fixed, remove this and use
	// ObjectMeta.Generation instead.
	// +optional
	Generation int64 `json:"generation,omitempty"`

	// ServingState holds a value describing the desired state the Kubernetes
	// resources should be in.
	ServingState ServingStateType `json:"servingState"`

	ScaleTargetRef autoscalev1.CrossVersionObjectReference `json:"scaleTargetRef"`
	// TODO(mattmoor): The actual spec.
}

// ServingStateType is an enumeration of the levels of serving readiness.
type ServingStateType string

const (
	ServingStateActive  ServingStateType = "Active"
	ServingStateReserve ServingStateType = "Reserve"
	ServingStateRetired ServingStateType = "Retired"
)

type PodAutoscalerCondition struct {
	Type PodAutoscalerConditionType `json:"type"`

	Status corev1.ConditionStatus `json:"status" description:"status of the condition, one of True, False, Unknown"`

	// +optional
	// TODO(mattmoor): Use VolatileTime (move to knative/pkg)
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" description:"last time the condition transit from one status to another"`

	// +optional
	Reason string `json:"reason,omitempty" description:"one-word CamelCase reason for the condition's last transition"`
	// +optional
	Message string `json:"message,omitempty" description:"human-readable message indicating details about last transition"`
}

// PodAutoscalerConditionType represents an PodAutoscaler condition value
type PodAutoscalerConditionType string

const (
	// PodAutoscalerConditionReady is set when the service is configured
	// and has available backends ready to receive traffic.
	PodAutoscalerConditionReady PodAutoscalerConditionType = "Ready"
)

type PodAutoscalerConditionSlice []PodAutoscalerCondition

// Len implements sort.Interface
func (scs PodAutoscalerConditionSlice) Len() int {
	return len(scs)
}

// Less implements sort.Interface
func (scs PodAutoscalerConditionSlice) Less(i, j int) bool {
	return scs[i].Type < scs[j].Type
}

// Swap implements sort.Interface
func (scs PodAutoscalerConditionSlice) Swap(i, j int) {
	scs[i], scs[j] = scs[j], scs[i]
}

var _ sort.Interface = (PodAutoscalerConditionSlice)(nil)

type PodAutoscalerStatus struct {
	// +optional
	Conditions PodAutoscalerConditionSlice `json:"conditions,omitempty"`

	// +optional
	ServingState ServingStateType `json:"servingState,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PodAutoscalerList is a list of PodAutoscaler resources
type PodAutoscalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []PodAutoscaler `json:"items"`
}

func (s *PodAutoscaler) GetGeneration() int64 {
	return s.Spec.Generation
}

func (s *PodAutoscaler) SetGeneration(generation int64) {
	s.Spec.Generation = generation
}

func (s *PodAutoscaler) GetSpecJSON() ([]byte, error) {
	return json.Marshal(s.Spec)
}

func (ss *PodAutoscalerStatus) IsReady() bool {
	if c := ss.GetCondition(PodAutoscalerConditionReady); c != nil {
		return c.Status == corev1.ConditionTrue
	}
	return false
}

func (ss *PodAutoscalerStatus) GetCondition(t PodAutoscalerConditionType) *PodAutoscalerCondition {
	for _, cond := range ss.Conditions {
		if cond.Type == t {
			return &cond
		}
	}
	return nil
}

// TODO(mattmoor): Condition stuff.
