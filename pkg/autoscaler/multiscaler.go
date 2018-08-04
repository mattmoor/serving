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
	"context"
	"fmt"
	"sync"
	"time"

	commonlogkey "github.com/knative/pkg/logging/logkey"
	kpav1alpha1 "github.com/knative/serving/pkg/apis/autoscaling/v1alpha1"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/knative/serving/pkg/logging"
	"github.com/knative/serving/pkg/logging/logkey"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
)

const (
	// Enough buffer to store scale requests generated every 2
	// seconds while an http request is taking the full timeout of 5
	// second.
	scaleBufferSize = 10
)

// UniScaler records statistics for a particular KPA and proposes the scale for the KPA based on those statistics.
type UniScaler interface {
	// Record records the given statistics.
	Record(context.Context, Stat)

	// Scale either proposes a number of replicas or skips proposing. The proposal is requested at the given time.
	// The returned boolean is true if and only if a proposal was returned.
	Scale(context.Context, time.Time) (int32, bool)
}

// UniScalerFactory creates a UniScaler for a given KPA using the given configuration.
type UniScalerFactory func(*kpav1alpha1.PodAutoscaler, *Config) (UniScaler, error)

// RevisionScaler knows how to scale revisions.
type RevisionScaler interface {
	// Scale attempts to scale the given revision to the desired scale.
	Scale(rev *v1alpha1.Revision, desiredScale int32)
}

// KPAScaler knows how to scale the target of a KPA.
type KPAScaler interface {
	// Scale attempts to scale the given KPA's target to the desired scale.
	Scale(kpa *kpav1alpha1.PodAutoscaler, desiredScale int32)
}

// scalerRunner wraps a UniScaler and a channel for implementing shutdown behavior.
type scalerRunner struct {
	scaler UniScaler
	stopCh chan struct{}
}

type key string

func newKey(namespace string, name string) key {
	return key(fmt.Sprintf("%s/%s", namespace, name))
}

// MultiScaler maintains a collection of UniScalers indexed by key.
type MultiScaler struct {
	scalers       map[key]*scalerRunner
	scalersMutex  sync.RWMutex
	scalersStopCh <-chan struct{}

	config *Config

	kpaScaler KPAScaler

	uniScalerFactory UniScalerFactory

	logger *zap.SugaredLogger
}

// NewMultiScaler constructs a MultiScaler.
func NewMultiScaler(config *Config, kpaScaler KPAScaler, stopCh <-chan struct{}, uniScalerFactory UniScalerFactory, logger *zap.SugaredLogger) *MultiScaler {
	logger.Debugf("Creating MultiScalar with configuration %#v", config)
	return &MultiScaler{
		scalers:          make(map[key]*scalerRunner),
		scalersStopCh:    stopCh,
		config:           config,
		kpaScaler:        kpaScaler,
		uniScalerFactory: uniScalerFactory,
		logger:           logger,
	}
}

// OnPresent adds, if necessary, a scaler for the given KPA.
func (m *MultiScaler) OnPresent(kpa *kpav1alpha1.PodAutoscaler, logger *zap.SugaredLogger) {
	m.scalersMutex.Lock()
	defer m.scalersMutex.Unlock()
	key := newKey(kpa.Namespace, kpa.Name)
	if _, exists := m.scalers[key]; !exists {
		ctx := logging.WithLogger(context.TODO(), logger)
		logger.Debug("Creating scaler for KPA.")
		scaler, err := m.createScaler(ctx, kpa)
		if err != nil {
			logger.Errorf("Failed to create scaler for KPA %#v: %v", kpa, err)
			return
		}
		logger.Info("Created scaler for KPA.")
		m.scalers[key] = scaler
	}
}

// OnAbsent removes, if necessary, a scaler for the KPA in the given namespace and with the given name.
func (m *MultiScaler) OnAbsent(namespace string, name string, logger *zap.SugaredLogger) {
	m.scalersMutex.Lock()
	defer m.scalersMutex.Unlock()
	key := newKey(namespace, name)
	if scaler, exists := m.scalers[key]; exists {
		close(scaler.stopCh)
		delete(m.scalers, key)
		logger.Info("Deleted scaler for KPA.")
	}
}

// loggerWithInfo enriches the logs with the KPA's name and namespace.
func loggerWithInfo(logger *zap.SugaredLogger, ns string, name string) *zap.SugaredLogger {
	return logger.With(zap.String(commonlogkey.Namespace, ns), zap.String(logkey.KPA, name))
}

func (m *MultiScaler) createScaler(ctx context.Context, kpa *kpav1alpha1.PodAutoscaler) (*scalerRunner, error) {
	scaler, err := m.uniScalerFactory(kpa, m.config)
	if err != nil {
		return nil, err
	}

	stopCh := make(chan struct{})
	runner := &scalerRunner{scaler: scaler, stopCh: stopCh}

	ticker := time.NewTicker(m.config.TickInterval)

	scaleChan := make(chan int32, scaleBufferSize)

	go func() {
		for {
			select {
			case <-m.scalersStopCh:
				ticker.Stop()
				return
			case <-stopCh:
				ticker.Stop()
				return
			case <-ticker.C:
				m.tickScaler(ctx, scaler, scaleChan)
			}
		}
	}()

	logger := logging.FromContext(ctx)

	go func() {
		for {
			select {
			case <-m.scalersStopCh:
				return
			case <-stopCh:
				return
			case desiredScale := <-scaleChan:
				m.kpaScaler.Scale(kpa, mostRecentDesiredScale(desiredScale, scaleChan, logger))
			}
		}
	}()

	return runner, nil
}

func mostRecentDesiredScale(desiredScale int32, scaleChan chan int32, logger *zap.SugaredLogger) int32 {
	for {
		select {
		case desiredScale = <-scaleChan:
			logger.Info("Scaling is not keeping up with autoscaling requests")
		default:
			// scaleChan is empty
			return desiredScale
		}
	}
}

func (m *MultiScaler) tickScaler(ctx context.Context, scaler UniScaler, scaleChan chan<- int32) {
	logger := logging.FromContext(ctx)
	desiredScale, scaled := scaler.Scale(ctx, time.Now())

	if scaled {
		// Cannot scale negative.
		if desiredScale < 0 {
			logger.Errorf("Cannot scale: desiredScale %d < 0.", desiredScale)
			return
		}

		// Don't scale to zero if scale to zero is disabled.
		if desiredScale == 0 && !m.config.EnableScaleToZero {
			logger.Warn("Cannot scale: Desired scale == 0 && EnableScaleToZero == false.")
			return
		}

		scaleChan <- desiredScale
	}
}

// RecordStat records some statistics for the given KPA. kpaKey should have the
// form namespace/name.
func (m *MultiScaler) RecordStat(kpaKey string, stat Stat) {
	m.scalersMutex.RLock()
	defer m.scalersMutex.RUnlock()

	scaler, exists := m.scalers[key(kpaKey)]
	if exists {
		namespace, name, err := cache.SplitMetaNamespaceKey(kpaKey)
		if err != nil {
			m.logger.Errorf("Invalid KPA key %s", kpaKey)
			return
		}
		logger := loggerWithInfo(m.logger, namespace, name)
		ctx := logging.WithLogger(context.TODO(), logger)
		scaler.scaler.Record(ctx, stat)
	}
}
