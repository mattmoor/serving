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

package autoscaler

import (
	kpav1alpha1 "github.com/knative/serving/pkg/apis/autoscaling/v1alpha1"
	clientset "github.com/knative/serving/pkg/client/clientset/versioned"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/scale"
	"strings"
)

// kpaScaler scales the target of a KPA up or down including scaling to zero.
type kpaScaler struct {
	servingClientSet clientset.Interface
	scaleClientSet   scale.ScalesGetter
	logger           *zap.SugaredLogger
}

// NewKPAScaler creates a kpaScaler.
func NewKPAScaler(servingClientSet clientset.Interface, scaleClientSet scale.ScalesGetter, logger *zap.SugaredLogger) KPAScaler {
	return &kpaScaler{
		servingClientSet: servingClientSet,
		scaleClientSet:   scaleClientSet,
		logger:           logger,
	}
}

// Scale attempts to scale the given KPA's target to the desired scale.
func (rs *kpaScaler) Scale(oldKPA *kpav1alpha1.PodAutoscaler, desiredScale int32) {
	// TODO(mattmoor): Add KPA context
	logger := rs.logger

	// Do not scale an inactive KPA
	// FIXME: given the input oldKPA is stale, it might be better to pass in the name and namespace instead.
	kpaClient := rs.servingClientSet.AutoscalingV1alpha1().PodAutoscalers(oldKPA.Namespace)
	kpa, err := kpaClient.Get(oldKPA.Name, metav1.GetOptions{})
	if err == nil && kpa.Spec.ServingState != kpav1alpha1.ServingStateActive {
		return
	}

	gv, err := schema.ParseGroupVersion(kpa.Spec.ScaleTargetRef.APIVersion)
	if err != nil {
		logger.Error("Unable to parse APIVersion.", zap.Error(err))
		return
	}
	resource := schema.GroupResource{
		Group: gv.Group,
		// TODO(mattmoor): Do something better than this.
		Resource: strings.ToLower(kpa.Spec.ScaleTargetRef.Kind) + "s",
	}
	scl, err := rs.scaleClientSet.Scales(kpa.Namespace).Get(resource, kpa.Spec.ScaleTargetRef.Name)
	if err != nil {
		logger.Error("ScaleTargetRef not found.", zap.Error(err))
		return
	}
	currentScale := scl.Spec.Replicas

	if desiredScale == currentScale {
		return
	}

	// Don't scale if current scale is zero. Rely on the activator to scale
	// from zero.
	if currentScale == 0 {
		logger.Warn("Cannot scale: Current scale is 0; activator must scale from 0.")
		return
	}
	logger.Infof("Scaling from %d to %d", currentScale, desiredScale)

	// When scaling to zero, flip the KPA's ServingState to Reserve.
	if desiredScale == 0 {
		logger.Debug("Setting KPA ServingState to Reserve.")
		kpa.Spec.ServingState = kpav1alpha1.ServingStateReserve
		if _, err := kpaClient.Update(kpa); err != nil {
			logger.Error("Error updating KPA serving state.", zap.Error(err))
		}
		return
	}

	// Scale the deployment.
	scl.Spec.Replicas = desiredScale
	_, err = rs.scaleClientSet.Scales(kpa.Namespace).Update(resource, scl)
	if err != nil {
		logger.Error("Unable to scale the ScaleTargetRef.", zap.Error(err))
		return
	}

	logger.Debug("Successfully scaled.")
}
