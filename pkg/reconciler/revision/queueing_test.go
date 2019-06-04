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

package revision

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/authn/k8schain"
	"golang.org/x/sync/errgroup"

	"github.com/knative/pkg/configmap"
	"github.com/knative/pkg/controller"
	"github.com/knative/pkg/logging"
	"github.com/knative/pkg/metrics"
	"github.com/knative/pkg/ptr"
	"github.com/knative/pkg/system"
	autoscalingv1alpha1 "github.com/knative/serving/pkg/apis/autoscaling/v1alpha1"
	"github.com/knative/serving/pkg/apis/serving"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/knative/serving/pkg/apis/serving/v1beta1"
	"github.com/knative/serving/pkg/autoscaler"
	fakeservingclient "github.com/knative/serving/pkg/client/injection/client/fake"
	"github.com/knative/serving/pkg/deployment"
	"github.com/knative/serving/pkg/network"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"

	. "github.com/knative/pkg/reconciler/testing"
	. "github.com/knative/serving/pkg/reconciler/testing"
)

type nopResolver struct{}

func (r *nopResolver) Resolve(_ string, _ k8schain.Options, _ sets.String) (string, error) {
	return "", nil
}

const (
	testAutoscalerImage            = "autoscalerImage"
	testFluentdImage               = "fluentdImage"
	testFluentdSidecarOutputConfig = `
<match **>
  @type elasticsearch
</match>
`
	testNamespace  = "test"
	testQueueImage = "queueImage"
)

func testRevision() *v1alpha1.Revision {
	return &v1alpha1.Revision{
		ObjectMeta: metav1.ObjectMeta{
			SelfLink:  "/apis/serving/v1alpha1/namespaces/test/revisions/test-rev",
			Name:      "test-rev",
			Namespace: testNamespace,
			Labels: map[string]string{
				"testLabel1":          "foo",
				"testLabel2":          "bar",
				serving.RouteLabelKey: "test-route",
			},
			Annotations: map[string]string{
				"testAnnotation": "test",
			},
			UID: "test-rev-uid",
		},
		Spec: v1alpha1.RevisionSpec{
			RevisionSpec: v1beta1.RevisionSpec{
				PodSpec: corev1.PodSpec{
					// corev1.Container has a lot of setting.  We try to pass many
					// of them here to verify that we pass through the settings to
					// derived objects.
					Containers: []corev1.Container{{
						Image:      "gcr.io/repo/image",
						Command:    []string{"echo"},
						Args:       []string{"hello", "world"},
						WorkingDir: "/tmp",
						Env: []corev1.EnvVar{{
							Name:  "EDITOR",
							Value: "emacs",
						}},
						LivenessProbe: &corev1.Probe{
							TimeoutSeconds: 42,
						},
						ReadinessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "health",
								},
							},
							TimeoutSeconds: 43,
						},
						TerminationMessagePath: "/dev/null",
					}},
				},
				TimeoutSeconds: ptr.Int64(60),
			},
		},
	}
}

func getTestDeploymentConfig() *deployment.Config {
	c, _ := deployment.NewConfigFromConfigMap(getTestDeploymentConfigMap())
	// ignoring error as test controller is generated
	return c
}

func getTestDeploymentConfigMap() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment.ConfigName,
			Namespace: system.Namespace(),
		},
		Data: map[string]string{
			"queueSidecarImage": testQueueImage,
			"autoscalerImage":   testAutoscalerImage,
		},
	}
}

func newTestController(t *testing.T) (
	context.Context,
	[]controller.Informer,
	*controller.Impl,
	*configmap.ManualWatcher) {

	ctx, informers := SetupFakeContext(t)
	configMapWatcher := &configmap.ManualWatcher{Namespace: system.Namespace()}
	controller := NewController(ctx, configMapWatcher)

	controller.Reconciler.(*Reconciler).resolver = &nopResolver{}

	configs := []*corev1.ConfigMap{
		getTestDeploymentConfigMap(),
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: system.Namespace(),
				Name:      network.ConfigName,
			}}, {
			ObjectMeta: metav1.ObjectMeta{
				Namespace: system.Namespace(),
				Name:      logging.ConfigMapName(),
			},
			Data: map[string]string{
				"zap-logger-config":   "{\"level\": \"error\",\n\"outputPaths\": [\"stdout\"],\n\"errorOutputPaths\": [\"stderr\"],\n\"encoding\": \"json\"}",
				"loglevel.queueproxy": "info",
			}}, {
			ObjectMeta: metav1.ObjectMeta{
				Namespace: system.Namespace(),
				Name:      metrics.ConfigMapName(),
			},
			Data: map[string]string{
				"logging.enable-var-log-collection":     "true",
				"logging.fluentd-sidecar-image":         testFluentdImage,
				"logging.fluentd-sidecar-output-config": testFluentdSidecarOutputConfig,
			}}, {
			ObjectMeta: metav1.ObjectMeta{
				Namespace: system.Namespace(),
				Name:      autoscaler.ConfigName,
			},
			Data: map[string]string{
				"max-scale-up-rate":                       "1.0",
				"container-concurrency-target-percentage": "0.5",
				"container-concurrency-target-default":    "10.0",
				"stable-window":                           "5m",
				"panic-window":                            "10s",
				"scale-to-zero-threshold":                 "10m",
				"tick-interval":                           "2s",
			}},
	}
	for _, configMap := range configs {
		configMapWatcher.OnChange(configMap)
	}

	return ctx, informers, controller, configMapWatcher
}

func TestNewRevisionCallsSyncHandler(t *testing.T) {
	ctx, informers, ctrl, _ := newTestController(t)
	ctx, cancel := context.WithCancel(ctx)
	eg := errgroup.Group{}
	defer func() {
		cancel()
		if err := eg.Wait(); err != nil {
			t.Fatalf("Error running controller: %v", err)
		}
	}()

	rev := testRevision()

	servingClient := fakeservingclient.Get(ctx)

	h := NewHooks()

	// Check for a service created as a signal that syncHandler ran
	h.OnCreate(&servingClient.Fake, "podautoscalers", func(obj runtime.Object) HookResult {
		pa := obj.(*autoscalingv1alpha1.PodAutoscaler)
		t.Logf("PA created: %s", pa.Name)
		return HookComplete
	})

	if err := controller.StartInformers(ctx.Done(), informers...); err != nil {
		t.Fatalf("Error starting informers: %v", err)
	}

	eg.Go(func() error {
		return ctrl.Run(2, ctx.Done())
	})

	if _, err := servingClient.ServingV1alpha1().Revisions(rev.Namespace).Create(rev); err != nil {
		t.Fatalf("Error creating revision: %v", err)
	}

	if err := h.WaitForHooks(time.Second * 3); err != nil {
		t.Error(err)
	}
}
