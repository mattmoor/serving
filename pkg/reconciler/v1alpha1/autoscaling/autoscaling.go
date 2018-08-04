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

package autoscaling

import (
	"context"
	"fmt"
	"time"

	"github.com/knative/pkg/controller"
	commonlogkey "github.com/knative/pkg/logging/logkey"
	kpav1alpha1 "github.com/knative/serving/pkg/apis/autoscaling/v1alpha1"
	informers "github.com/knative/serving/pkg/client/informers/externalversions/autoscaling/v1alpha1"
	listers "github.com/knative/serving/pkg/client/listers/autoscaling/v1alpha1"
	"github.com/knative/serving/pkg/logging/logkey"
	"github.com/knative/serving/pkg/reconciler"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
)

const (
	controllerAgentName = "autoscaling-controller"
)

// KPASynchronizer is an interface for notifying the presence or absence of KPAs.
type KPASynchronizer interface {
	// OnPresent is called when the given KPA exists.
	OnPresent(rev *kpav1alpha1.PodAutoscaler, logger *zap.SugaredLogger)

	// OnAbsent is called when a KPA in the given namespace with the given name ceases to exist.
	OnAbsent(namespace string, name string, logger *zap.SugaredLogger)
}

// Reconciler tracks KPAs and notifies a KPASynchronizer of their presence and absence.
type Reconciler struct {
	*reconciler.Base

	kpaLister listers.PodAutoscalerLister
	kpaSynch  KPASynchronizer
}

// Check that our Reconciler implements controller.Reconciler
var _ controller.Reconciler = (*Reconciler)(nil)

// NewController creates an autoscaling Controller.
func NewController(
	opts *reconciler.Options,
	kpaInformer informers.PodAutoscalerInformer,
	kpaSynch KPASynchronizer,
	informerResyncInterval time.Duration,
) *controller.Impl {

	c := &Reconciler{
		Base:      reconciler.NewBase(*opts, controllerAgentName),
		kpaLister: kpaInformer.Lister(),
		kpaSynch:  kpaSynch,
	}
	impl := controller.NewImpl(c, c.Logger, "Autoscaling")

	c.Logger.Info("Setting up event handlers")
	kpaInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    impl.Enqueue,
		UpdateFunc: controller.PassNew(impl.Enqueue),
		DeleteFunc: impl.Enqueue,
	})

	return impl
}

// Reconcile notifies the KPASynchronizer of the presence or absence.
func (c *Reconciler) Reconcile(ctx context.Context, key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key %s: %v", key, err))
		return nil
	}

	logger := loggerWithKPAInfo(c.Logger, namespace, name)
	logger.Debug("Reconcile KPA")

	kpa, err := c.kpaLister.PodAutoscalers(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Debug("KPA no longer exists")
			c.kpaSynch.OnAbsent(namespace, name, logger)
			return nil
		}
		runtime.HandleError(err)
		return err
	}

	logger.Debug("KPA exists")
	c.kpaSynch.OnPresent(kpa, logger)

	// TODO(mattmoor): Update the KPA Status.

	return nil
}

func loggerWithKPAInfo(logger *zap.SugaredLogger, ns string, name string) *zap.SugaredLogger {
	return logger.With(zap.String(commonlogkey.Namespace, ns), zap.String(logkey.KPA, name))
}
