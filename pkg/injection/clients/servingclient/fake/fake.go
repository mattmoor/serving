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

package fake

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"

	"github.com/knative/pkg/injection"
	"github.com/knative/serving/pkg/client/clientset/versioned/fake"
	"github.com/knative/serving/pkg/injection/clients/servingclient"
)

func init() {
	injection.Fake.RegisterClient(withClient)
}

func withClient(ctx context.Context, cfg *rest.Config) context.Context {
	ctx, _ = With(ctx)
	return ctx
}

func With(ctx context.Context, objects ...runtime.Object) (context.Context, *fake.Clientset) {
	cs := fake.NewSimpleClientset(objects...)
	return context.WithValue(ctx, servingclient.Key{}, cs), cs
}

// Get extracts the certmanager client from the context.
func Get(ctx context.Context) *fake.Clientset {
	return ctx.Value(servingclient.Key{}).(*fake.Clientset)
}
