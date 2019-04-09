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

	"github.com/knative/pkg/apis"
	"github.com/knative/serving/pkg/apis/serving"
)

// Validate makes sure that Configuration is properly configured.
func (c *Configuration) Validate(ctx context.Context) *apis.FieldError {
	// TODO(mattmoor): Add a context for passing in the parent object's name.
	return serving.ValidateObjectMetadata(c.GetObjectMeta()).ViaField("metadata").Also(
		c.Spec.Validate(withinSpec(ctx)).ViaField("spec")).Also(
		c.Status.Validate(withinStatus(ctx)).ViaField("status"))
}

// Validate implements apis.Validatable
func (cs *ConfigurationSpec) Validate(ctx context.Context) *apis.FieldError {
	return cs.Template.Validate(ctx).ViaField("template")
}

// Validate implements apis.Validatable
func (cs *ConfigurationStatus) Validate(ctx context.Context) *apis.FieldError {
	return cs.ConfigurationStatusFields.Validate(ctx)
}

// Validate implements apis.Validatable
func (csf *ConfigurationStatusFields) Validate(ctx context.Context) *apis.FieldError {
	return nil
}
