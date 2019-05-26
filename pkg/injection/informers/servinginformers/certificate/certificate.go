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

package certificate

import (
	"context"

	networkinginformers "github.com/knative/serving/pkg/client/informers/externalversions/networking/v1alpha1"

	"github.com/knative/pkg/controller"
	"github.com/knative/pkg/injection"
	"github.com/knative/serving/pkg/injection/informers/servinginformers/factory"
)

func init() {
	injection.Default.RegisterInformer(withCertificateInformer)
}

// Key is used as the key for associating information
// with a context.Context.
type Key struct{}

func withCertificateInformer(ctx context.Context) (context.Context, controller.Informer) {
	f := factory.Get(ctx)
	inf := f.Networking().V1alpha1().Certificates()
	return context.WithValue(ctx, Key{}, inf), inf.Informer()
}

// Get extracts the Knative Certificate informer from the context.
func Get(ctx context.Context) networkinginformers.CertificateInformer {
	return ctx.Value(Key{}).(networkinginformers.CertificateInformer)
}
