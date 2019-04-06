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

package v1beta1

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/knative/pkg/apis"
	netv1alpha1 "github.com/knative/serving/pkg/apis/networking/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func TestContainerConcurrencyValidation(t *testing.T) {
	tests := []struct {
		name string
		cc   RevisionContainerConcurrencyType
		want *apis.FieldError
	}{{
		name: "single",
		cc:   1,
		want: nil,
	}, {
		name: "unlimited",
		cc:   0,
		want: nil,
	}, {
		name: "ten",
		cc:   10,
		want: nil,
	}, {
		name: "invalid container concurrency (too small)",
		cc:   -1,
		want: apis.ErrOutOfBoundsValue("-1", "0", revisionContainerConcurrencyMax,
			apis.CurrentField),
	}, {
		name: "invalid container concurrency (too large)",
		cc:   RevisionContainerConcurrencyMax + 1,
		want: apis.ErrOutOfBoundsValue(
			fmt.Sprintf("%d", int(RevisionContainerConcurrencyMax)+1),
			"0", revisionContainerConcurrencyMax, apis.CurrentField),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.cc.Validate(context.Background())
			if diff := cmp.Diff(test.want.Error(), got.Error()); diff != "" {
				t.Errorf("Validate (-want, +got) = %v", diff)
			}
		})
	}
}

func TestRevisionSpecValidation(t *testing.T) {
	tests := []struct {
		name string
		rs   *RevisionSpec
		want *apis.FieldError
	}{{
		name: "valid",
		rs: &RevisionSpec{
			PodSpec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Image: "helloworld",
				}},
			},
		},
		want: nil,
	}, {
		name: "with volume (ok)",
		rs: &RevisionSpec{
			PodSpec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Image: "helloworld",
					VolumeMounts: []corev1.VolumeMount{{
						MountPath: "/mount/path",
						Name:      "the-name",
						ReadOnly:  true,
					}},
				}},
				Volumes: []corev1.Volume{{
					Name: "the-name",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "foo",
						},
					},
				}},
			},
		},
		want: nil,
		// TODO(mattmoor): Validate PodSpec
		// }, {
		// 	name: "with volume name collision",
		// 	rs: &RevisionSpec{
		// 		PodSpec: corev1.PodSpec{
		// 			Containers: []corev1.Container{{
		// 				Image: "helloworld",
		// 				VolumeMounts: []corev1.VolumeMount{{
		// 					MountPath: "/mount/path",
		// 					Name:      "the-name",
		// 					ReadOnly:  true,
		// 				}},
		// 			}},
		// 			Volumes: []corev1.Volume{{
		// 				Name: "the-name",
		// 				VolumeSource: corev1.VolumeSource{
		// 					Secret: &corev1.SecretVolumeSource{
		// 						SecretName: "foo",
		// 					},
		// 				},
		// 			}, {
		// 				Name: "the-name",
		// 				VolumeSource: corev1.VolumeSource{
		// 					ConfigMap: &corev1.ConfigMapVolumeSource{},
		// 				},
		// 			}},
		// 		},
		// 	},
		// 	want: (&apis.FieldError{
		// 		Message: fmt.Sprintf(`duplicate volume name "the-name"`),
		// 		Paths:   []string{"name"},
		// 	}).ViaFieldIndex("volumes", 1),
		// }, {
		// 	name: "bad pod spec",
		// 	rs: &RevisionSpec{
		// 		PodSpec: corev1.PodSpec{
		// 			Containers: []corev1.Container{{
		// 				Name:  "steve",
		// 				Image: "helloworld",
		// 			}},
		// 		},
		// 	},
		// 	want: apis.ErrDisallowedFields("container.name"),
	}, {
		name: "exceed max timeout",
		rs: &RevisionSpec{
			PodSpec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Image: "helloworld",
				}},
			},
			TimeoutSeconds: 6000,
		},
		want: apis.ErrOutOfBoundsValue("6000s", "0s",
			fmt.Sprintf("%ds", int(netv1alpha1.DefaultTimeout.Seconds())),
			"timeoutSeconds"),
	}, {
		name: "negative timeout",
		rs: &RevisionSpec{
			PodSpec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Image: "helloworld",
				}},
			},
			TimeoutSeconds: -30,
		},
		want: apis.ErrOutOfBoundsValue("-30s", "0s",
			fmt.Sprintf("%ds", int(netv1alpha1.DefaultTimeout.Seconds())),
			"timeoutSeconds"),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.rs.Validate(context.Background())
			if diff := cmp.Diff(test.want.Error(), got.Error()); diff != "" {
				t.Errorf("Validate (-want, +got) = %v", diff)
			}
		})
	}
}
