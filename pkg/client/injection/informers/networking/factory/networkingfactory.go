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

// Code generated by injection-gen. DO NOT EDIT.

package networkingfactory

import (
	"context"

	controller "github.com/knative/pkg/controller"
	injection "github.com/knative/pkg/injection"
	externalversions "github.com/knative/serving/pkg/client/informers/externalversions"
	client "github.com/knative/serving/pkg/client/injection/client"
)

func init() {
	injection.Default.RegisterInformerFactory(withInformerFactory)
}

// key is used as the key for associating information with a context.Context.
type Key struct{}

func withInformerFactory(ctx context.Context) context.Context {
	c := client.Get(ctx)
	return context.WithValue(ctx, Key{},
		externalversions.NewSharedInformerFactory(c, controller.GetResyncPeriod(ctx)))
}

// Get extracts the InformerFactory from the context.
func Get(ctx context.Context) externalversions.SharedInformerFactory {
	untyped := ctx.Value(Key{})
	if untyped == nil {
		return nil
	}
	return untyped.(externalversions.SharedInformerFactory)
}
