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

	"github.com/knative/pkg/apis"
	networkingv1alpha1 "github.com/knative/serving/pkg/apis/networking/v1alpha1"
	"github.com/knative/serving/pkg/apis/serving"
)

// Validate ensures Revision is properly configured.
func (r *Revision) Validate(ctx context.Context) *apis.FieldError {
	return serving.ValidateObjectMetadata(r.GetObjectMeta()).ViaField("metadata").
		Also(r.Spec.Validate(ctx).ViaField("spec"))
}

// Validate implements apis.Validatable.
func (rs *RevisionSpec) Validate(ctx context.Context) *apis.FieldError {
	err := rs.ContainerConcurrency.Validate(ctx).ViaField("containerConcurrency")

	// TODO(mattmoor): Validate pod spec.

	// TODO(mattmoor): Move this out of v1alpha1!
	max := int64(networkingv1alpha1.DefaultTimeout.Seconds())
	if rs.TimeoutSeconds < 0 || rs.TimeoutSeconds > max {
		err = err.Also(apis.ErrOutOfBoundsValue(fmt.Sprintf("%ds", rs.TimeoutSeconds),
			"0s", fmt.Sprintf("%ds", max), "timeoutSeconds"))
	}

	return err
}

// Validate implements apis.Validatable.
func (cc RevisionContainerConcurrencyType) Validate(ctx context.Context) *apis.FieldError {
	if cc < 0 || cc > RevisionContainerConcurrencyMax {
		return apis.ErrOutOfBoundsValue(fmt.Sprintf("%d", cc),
			"0", revisionContainerConcurrencyMax, apis.CurrentField)
	}
	return nil
}

// CheckImmutableFields checks the immutable fields are not modified.
func (current *Revision) CheckImmutableFields(ctx context.Context, og apis.Immutable) *apis.FieldError {
	return nil
}
