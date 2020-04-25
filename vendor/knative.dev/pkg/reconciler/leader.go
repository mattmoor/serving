/*
Copyright 2020 The Knative Authors

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

package reconciler

import (
	"sync"

	"k8s.io/apimachinery/pkg/types"
)

// LeaderAware is implemented by Reconcilers that are aware of their leader status.
type LeaderAware interface {
	// IsLeader returns the current leader status of the reconciler.
	IsLeader() bool

	// Promote is called when we become the leader.  It must be supplied with an
	// enqueue function through which a global resync may be triggered.
	Promote(enq func(types.NamespacedName))

	// Demote is called when we lose the leader election.
	Demote()
}

// LeaderAwareFuncs implements LeaderAware using the given functions for handling
// promotion and demotion.
type LeaderAwareFuncs struct {
	sync.RWMutex
	leader bool

	PromoteFunc func(enq func(types.NamespacedName))
	DemoteFunc  func()
}

var _ LeaderAware = (*LeaderAwareFuncs)(nil)

// IsLeader implements LeaderAware
func (laf *LeaderAwareFuncs) IsLeader() bool {
	laf.RLock()
	defer laf.RUnlock()
	return laf.leader
}

// Promote implements LeaderAware
func (laf *LeaderAwareFuncs) Promote(enq func(types.NamespacedName)) {
	promote := func() func(func(types.NamespacedName)) {
		laf.Lock()
		defer laf.Unlock()
		laf.leader = true
		return laf.PromoteFunc
	}()

	if promote != nil {
		promote(enq)
	}
}

// Demote implements LeaderAware
func (laf *LeaderAwareFuncs) Demote() {
	demote := func() func() {
		laf.Lock()
		defer laf.Unlock()
		laf.leader = false
		return laf.DemoteFunc
	}()

	if demote != nil {
		demote()
	}
}
