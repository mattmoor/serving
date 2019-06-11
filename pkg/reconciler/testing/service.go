/*
Copyright 2019 The Knative Authors

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

package testing

import (
	"context"

	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/knative/serving/pkg/apis/serving/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LegacyService creates a service with LegacyServiceOptions
func LegacyService(name, namespace string, so ...LegacyServiceOption) *v1alpha1.Service {
	s := &v1alpha1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	for _, opt := range so {
		opt(s)
	}
	s.SetDefaults(context.Background())
	return s
}

// LegacyServiceWithoutNamespace creates a service with LegacyServiceOptions but without a specific namespace
func LegacyServiceWithoutNamespace(name string, so ...LegacyServiceOption) *v1alpha1.Service {
	s := &v1alpha1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	for _, opt := range so {
		opt(s)
	}
	s.SetDefaults(context.Background())
	return s
}

// Service creates a service with ServiceOptions
func Service(name, namespace string, so ...ServiceOption) *v1beta1.Service {
	s := &v1beta1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	for _, opt := range so {
		opt(s)
	}
	s.SetDefaults(context.Background())
	return s
}

// ServiceWithoutNamespace creates a service with ServiceOptions but without a specific namespace
func ServiceWithoutNamespace(name string, so ...ServiceOption) *v1beta1.Service {
	s := &v1beta1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	for _, opt := range so {
		opt(s)
	}
	s.SetDefaults(context.Background())
	return s
}
