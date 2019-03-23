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

// Versionable handles conversions to/from types of a lower version.
// The receiver is always the higher version, and the argument is always the lower version.
// Depending on the direction of the conversion (upgrade or downgrade) the receiver may be
// the source or the sink of the transformation.
type Versionable interface {
	// UpFrom converts from the provided Versionable into the receiver.
	UpFrom(from Versionable) error

	// DownTo converts the receiver to the provided Versionable.
	DownTo(to Versionable) error
}
