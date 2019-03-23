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

package v1beta1

import (
	"context"
	"errors"

	"github.com/knative/pkg/apis"
	duckv1alpha1 "github.com/knative/pkg/apis/duck/v1alpha1"
	"github.com/knative/pkg/kmeta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Service acts as a top-level container that manages a Route and Configuration
// which implement a network service. Service exists to provide a singular
// abstraction which can be access controlled, reasoned about, and which
// encapsulates software lifecycle decisions such as rollout policy and
// team resource ownership. Service acts only as an orchestrator of the
// underlying Routes and Configurations (much as a kubernetes Deployment
// orchestrates ReplicaSets), and its usage is optional but recommended.
//
// The Service's controller will track the statuses of its owned Configuration
// and Route, reflecting their statuses and conditions as its own.
//
// See also: https://github.com/knative/serving/blob/master/docs/spec/overview.md#service
type Service struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	Spec ServiceSpec `json:"spec,omitempty"`

	// +optional
	Status ServiceStatus `json:"status,omitempty"`
}

// Verify that Service adheres to the appropriate interfaces.
var (
	// Check that Service may be validated and defaulted.
	_ apis.Validatable = (*Service)(nil)
	_ apis.Defaultable = (*Service)(nil)
	_ apis.Versionable = (*Service)(nil)

	// Check that we can create OwnerReferences to a Service.
	_ kmeta.OwnerRefable = (*Service)(nil)
)

// ServiceSpec represents the configuration for the Service object. Exactly one
// of its members (other than Generation) must be specified. Services can either
// track the latest ready revision of a configuration or be pinned to a specific
// revision.
type ServiceSpec struct {
	// ServiceSpec inlines an unrestricted ConfigurationSpec.
	ConfigurationSpec `json:",inline"`

	// ServiceSpec inlines RouteSpec and restricts/defaults its fields
	// via webhook.  In particular, this spec can only reference this
	// Service's configuration and revisions (which also influences
	// defaults).
	RouteSpec `json:",inline"`
}

// TODO(mattmoor): DO NOT SUBMIT
// TODO(mattmoor): Verify that Service implements PodSpec-able

// ConditionType represents a Service condition value
const (
	// ServiceConditionReady is set when the service is configured
	// and has available backends ready to receive traffic.
	ServiceConditionReady = duckv1alpha1.ConditionReady
	// ServiceConditionRouteReady is set when the service's underlying
	// routes have reported readiness.
	ServiceConditionRouteReady duckv1alpha1.ConditionType = "RouteReady"
	// ServiceConditionConfigurationReady is set when the service's underlying
	// configurations have reported readiness.
	ServiceConditionConfigurationReady duckv1alpha1.ConditionType = "ConfigurationReady"
)

// ServiceStatus represents the Status stanza of the Service resource.
type ServiceStatus struct {
	duckv1alpha1.Status `json:",inline"`

	// In addition to inlining ConfigurationSpec, we also inline the fields
	// specific to ConfigurationStatus.
	ConfigurationStatusFields `json:",inline"`

	// In addition to inlining RouteSpec, we also inline the fields
	// specific to RouteStatus.
	RouteStatusFields `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ServiceList is a list of Service resources
type ServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Service `json:"items"`
}

func (r *Service) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Service")
}

func (c *Service) Validate(ctx context.Context) *apis.FieldError {
	return nil
}

func (c *Service) SetDefaults(ctx context.Context) {
}

// UpFrom populates the receiver with the up-converted input object.
func (s *Service) UpFrom(obj apis.Versionable) error {
	s.APIVersion, s.Kind = s.GetGroupVersionKind().ToAPIVersionAndKind()
	switch src := obj.(type) {
	case *v1alpha1.Service:
		s.ObjectMeta = src.ObjectMeta
		if err := s.Spec.UpFrom(src.Spec); err != nil {
			return err
		}
		if err := s.Status.UpFrom(src.Status); err != nil {
			return err
		}
		return nil
	default:
		return errors.New("Unrecognized source type.")
	}
}

// UpFrom populates the receiver with the up-converted input object.
func (s *ServiceSpec) UpFrom(src v1alpha1.ServiceSpec) error {
	switch {
	case src.RunLatest != nil:
		// TODO(mattmoor): Populate the RouteSpec
		return s.ConfigurationSpec.UpFrom(src.RunLatest.Configuration)
	case src.Release != nil:
		return errors.New("TODO: Release mode conversion.")
	default:
		return errors.New("Unsupported mode.")
	}
}

// UpFrom populates the receiver with the up-converted input object.
func (s *ServiceStatus) UpFrom(src v1alpha1.ServiceStatus) error {
	// TODO(mattmoor): Translate conditions.
	s.Status = src.Status

	s.ConfigurationStatusFields.LatestReadyRevisionName = src.LatestReadyRevisionName
	s.ConfigurationStatusFields.LatestCreatedRevisionName = src.LatestCreatedRevisionName

	s.RouteStatusFields.Domain = src.Domain
	s.RouteStatusFields.Address = src.Address
	for _, t := range src.Traffic {
		s.RouteStatusFields.Traffic = append(s.RouteStatusFields.Traffic, TrafficTarget{
			Name:              t.Name,
			RevisionName:      t.RevisionName,
			ConfigurationName: t.ConfigurationName,
			Percent:           t.Percent,
			// TODO(mattmoor): url
		})
	}
	return nil
}

// DownTo populates the provided input object with the appropriately
// down-converted object.
func (s *Service) DownTo(obj apis.Versionable) error {
	return errors.New("TODO: Implement down-conversion.")
}
