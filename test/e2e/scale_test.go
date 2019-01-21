// +build e2e

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

package e2e

import (
	"testing"
	"time"

	"github.com/knative/pkg/test/logging"
)

type nopLatencies struct {
	logger *logging.BaseLogger
}

var _ Latencies = (*nopLatencies)(nil)

func (nl *nopLatencies) Add(metric string, start time.Time) {
	duration := time.Since(start)

	nl.logger.Infof("%q took %v", metric, duration)
}

// While redundant, we run two versions of this by default:
// 1. TestScaleTo10: a developer smoke test that's useful when changing this to assess whether
//   things have gone horribly wrong.  This should take about 12-20 seconds total.
// 2. TestScaleTo50: a more proper execution of the test, which verifies a slightly more
//   interesting burst of deployments, but low enough to complete in a reasonable window.

func TestScaleTo10(t *testing.T) {
	// Add test case specific name to its own logger.
	logger := logging.GetContextLogger(t.Name())

	ScaleToWithin(t, logger, 10, 60*time.Second, &nopLatencies{logger})
}

func TestScaleTo50(t *testing.T) {
	// Add test case specific name to its own logger.
	logger := logging.GetContextLogger(t.Name())

	ScaleToWithin(t, logger, 50, 5*time.Minute, &nopLatencies{logger})
}

// A version to customize for more extreme scale testing.
// This should only be checked in commented out.
// func TestScaleToN(t *testing.T) {
// 	// Add test case specific name to its own logger.
// 	logger := logging.GetContextLogger(t.Name())
//
// 	testScaleToWithin(t, logger, 200, 20*time.Minute, &nopLatencies{})
// }
