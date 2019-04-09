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

package v1alpha1

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	duckv1beta1 "github.com/knative/pkg/apis/duck/v1beta1"
	"github.com/knative/serving/pkg/apis/serving/v1beta1"
)

func TestServiceConversionBadType(t *testing.T) {
	good, bad := &Service{}, &Revision{}

	if err := good.ConvertUp(context.Background(), bad); err == nil {
		t.Errorf("ConvertUp() = %#v, wanted error", bad)
	}

	if err := good.ConvertDown(context.Background(), bad); err == nil {
		t.Errorf("ConvertDown() = %#v, wanted error", good)
	}
}

func TestServiceConversion(t *testing.T) {
	// TODO(mattmoor): This.
	t.Skip("ConvertDown(ServiceSpec)")

	tests := []struct {
		name string
		in   *Service
	}{{
		name: "config name",
		in: &Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:       "asdf",
				Namespace:  "blah",
				Generation: 1,
			},
			Spec: ServiceSpec{
				RunLatest: &RunLatestType{
					Configuration: ConfigurationSpec{},
				},
			},
			Status: ServiceStatus{
				Status: duckv1beta1.Status{
					ObservedGeneration: 1,
					Conditions: duckv1beta1.Conditions{{
						Type:   "Ready",
						Status: "True",
					}},
				},
			},
		},
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			beta := &v1beta1.Service{}
			if err := test.in.ConvertUp(context.Background(), beta); err != nil {
				t.Errorf("ConvertUp() = %v", err)
			}
			t.Logf("ConvertUp() = %#v", beta)
			got := &Service{}
			if err := got.ConvertDown(context.Background(), beta); err != nil {
				t.Errorf("ConvertDown() = %v", err)
			}
			t.Logf("ConvertDown() = %#v", got)
			if diff := cmp.Diff(test.in, got); diff != "" {
				t.Errorf("roundtrip (-want, +got) = %v", diff)
			}
		})
	}
}
