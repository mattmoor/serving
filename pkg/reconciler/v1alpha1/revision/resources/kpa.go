/*
Copyright 2018 The Knative Authors

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

package resources

import (
	autoscalev1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kpav1alpha1 "github.com/knative/serving/pkg/apis/autoscaling/v1alpha1"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/knative/serving/pkg/reconciler"
	"github.com/knative/serving/pkg/reconciler/v1alpha1/revision/resources/names"
)

func MakeKPA(rev *v1alpha1.Revision) *kpav1alpha1.PodAutoscaler {
	return &kpav1alpha1.PodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:            names.KPA(rev),
			Namespace:       rev.Namespace,
			Labels:          makeLabels(rev),
			Annotations:     makeAnnotations(rev),
			OwnerReferences: []metav1.OwnerReference{*reconciler.NewControllerRef(rev)},
		},
		Spec: kpav1alpha1.PodAutoscalerSpec{
			// TODO(mattmoor): Reconcile these two copied of ServingState types.
			ServingState: kpav1alpha1.ServingStateType(rev.Spec.ServingState),
			ScaleTargetRef: autoscalev1.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       names.Deployment(rev),
			},
		},
	}
}
