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

package apis

import (
	"context"

	authenticationv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Defaultable defines an interface for setting the defaults for the
// uninitialized fields of this instance.
type Defaultable interface {
	SetDefaults(context.Context)
}

// Validatable indicates that a particular type may have its fields validated.
type Validatable interface {
	// Validate checks the validity of this types fields.
	Validate(context.Context) *FieldError
}

// Immutable indicates that a particular type has fields that should
// not change after creation.
type Immutable interface {
	// CheckImmutableFields checks that the current instance's immutable
	// fields haven't changed from the provided original.
	CheckImmutableFields(ctx context.Context, original Immutable) *FieldError
}

// Listable indicates that a particular type can be returned via the returned
// list type by the API server.
type Listable interface {
	runtime.Object

	GetListType() runtime.Object
}

// Annotatable indicates that a particular type applies various annotations.
type Annotatable interface {
	AnnotateUserInfo(ctx context.Context, previous Annotatable, ui *authenticationv1.UserInfo)
}

// Convertible handles conversions to/from types of a higher version.
// The receiver is always the lower version, and the argument is always the
// higher version.  Depending on the direction of the conversion (upgrade or
// downgrade) the receiver may be the source or the sink of the transformation.
// We handle both directions of conversion here to avoid dependency cycles
// between high and low API version types.  We take the unidirectional
// dependency in this direction so that as lower versions are dropped the
// colocated conversions disappear alongside them, keeping the higher level
// API clean.
type Convertible interface {
	// DownFrom downgrades from the provided Convertible into the receiver.
	DownFrom(from Convertible) error

	// UpTo upgrades from the receiver into the provided Convertible.
	UpTo(to Convertible) error

	// IsStorage indicates whether this Convertible's form should be used
	// for storage.  Only one Version within a GroupKind may return true.
	IsStorage() bool
}
