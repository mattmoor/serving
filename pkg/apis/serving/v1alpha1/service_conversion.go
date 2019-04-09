/*
Copyright 2019 The Knative Authors.

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
	"fmt"

	"github.com/knative/pkg/apis"
	"github.com/knative/serving/pkg/apis/serving/v1beta1"
)

// ConvertUp implements apis.Convertible
func (source *Service) ConvertUp(ctx context.Context, obj apis.Convertible) error {
	switch sink := obj.(type) {
	case *v1beta1.Service:
		sink.ObjectMeta = source.ObjectMeta
		if err := source.Spec.ConvertUp(ctx, &sink.Spec); err != nil {
			return err
		}
		return source.Status.ConvertUp(ctx, &sink.Status)
	default:
		return fmt.Errorf("unknown version, got: %T", sink)
	}
}

// ConvertUp helps implement apis.Convertible
func (source *ServiceSpec) ConvertUp(ctx context.Context, sink *v1beta1.ServiceSpec) error {
	switch {
	case source.RunLatest != nil:
		sink.RouteSpec = v1beta1.RouteSpec{
			Traffic: []v1beta1.TrafficTarget{{
				Percent: 100,
			}},
		}
		return source.RunLatest.Configuration.ConvertUp(ctx, &sink.ConfigurationSpec)

	case source.Release != nil:
		if len(source.Release.Revisions) == 2 {
			sink.RouteSpec = v1beta1.RouteSpec{
				Traffic: []v1beta1.TrafficTarget{{
					RevisionName: source.Release.Revisions[0],
					Percent:      100 - source.Release.RolloutPercent,
					Subroute:     "current",
				}, {
					RevisionName: source.Release.Revisions[1],
					Percent:      source.Release.RolloutPercent,
					Subroute:     "candidate",
				}, {
					Percent:  0,
					Subroute: "latest",
				}},
			}
		} else {
			sink.RouteSpec = v1beta1.RouteSpec{
				Traffic: []v1beta1.TrafficTarget{{
					RevisionName: source.Release.Revisions[0],
					Percent:      100,
					Subroute:     "current",
				}, {
					Percent:  0,
					Subroute: "latest",
				}},
			}
		}
		for i, tt := range sink.RouteSpec.Traffic {
			if tt.RevisionName == "@latest" {
				sink.RouteSpec.Traffic[i].RevisionName = ""
			}
		}
		return source.Release.Configuration.ConvertUp(ctx, &sink.ConfigurationSpec)

	case source.DeprecatedPinned != nil:
		sink.RouteSpec = v1beta1.RouteSpec{
			Traffic: []v1beta1.TrafficTarget{{
				RevisionName: source.DeprecatedPinned.RevisionName,
				Percent:      100,
			}},
		}
		return source.Release.Configuration.ConvertUp(ctx, &sink.ConfigurationSpec)

	case source.Manual != nil:
		return ConvertErrorf("manual", "manual mode cannot be migrated forward, got %#v", source)

	default:
		return fmt.Errorf("unknown service mode %#v", source)
	}
}

// ConvertUp helps implement apis.Convertible
func (source *ServiceStatus) ConvertUp(ctx context.Context, sink *v1beta1.ServiceStatus) error {
	source.Status.ConvertTo(ctx, &sink.Status)

	source.RouteStatusFields.ConvertUp(ctx, &sink.RouteStatusFields)
	return source.ConfigurationStatusFields.ConvertUp(ctx, &sink.ConfigurationStatusFields)
}

// ConvertDown implements apis.Convertible
func (sink *Service) ConvertDown(ctx context.Context, obj apis.Convertible) error {
	switch source := obj.(type) {
	case *v1beta1.Service:
		sink.ObjectMeta = source.ObjectMeta
		if err := sink.Spec.ConvertDown(ctx, source.Spec); err != nil {
			return err
		}
		return sink.Status.ConvertDown(ctx, source.Status)
	default:
		return fmt.Errorf("unknown version, got: %T", source)
	}
}

// ConvertDown helps implement apis.Convertible
func (sink *ServiceSpec) ConvertDown(ctx context.Context, source v1beta1.ServiceSpec) error {
	// TODO(mattmoor): Consider waiting for when we inline v1beta1.RouteSpec?
	return nil
}

// ConvertDown helps implement apis.Convertible
func (sink *ServiceStatus) ConvertDown(ctx context.Context, source v1beta1.ServiceStatus) error {
	source.Status.ConvertTo(ctx, &sink.Status)

	sink.RouteStatusFields.ConvertDown(ctx, source.RouteStatusFields)
	return sink.ConfigurationStatusFields.ConvertDown(ctx, source.ConfigurationStatusFields)
}
