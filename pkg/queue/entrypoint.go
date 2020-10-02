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

package queue

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/cgi"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/util/sets"
	"knative.dev/pkg/network/handlers"
)

// EntrypointPrefix is the executable prefix used when
// invoking the Queue in entrypoint mode.
const EntrypointPrefix = "/knative"

var (
	// EntrypointModes contains the list of valid entrypoint modes.
	EntrypointModes = sets.NewString()

	modes = map[string]func() http.Handler{
		"cgi":        makeCGI,
		"entrypoint": makeCGI,
	}
)

func init() {
	for key := range modes {
		EntrypointModes.Insert(key)
	}
}

// Check that the binary exists and that it is executable.
// This is to guard against common error cases.
func checkValidBinary(binary string) error {
	stat, err := os.Stat(binary)
	if err != nil {
		return fmt.Errorf("Unable to stat %q: %v", binary, err)
	}
	if stat.Mode()&0111 == 0 {
		return fmt.Errorf("%q is not executable", binary)
	}
	return nil
}

// https://tools.ietf.org/html/rfc3875
func makeCGI() http.Handler {
	if len(os.Args) == 1 {
		log.Fatal("Not enough arguments to: ", os.Args[0])
	}
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal("Unable to determine working directory: ", err)
	}
	binary := os.Args[1]
	if err := checkValidBinary(binary); err != nil {
		log.Fatal(err.Error())
	}

	return &cgi.Handler{
		Path: binary,
		Args: os.Args[2:],
		Dir:  wd,
		// TODO(mattmoor): Do we need to pass through env?
	}
}

// IsEntrypoint checks whether the queue has been invoked in an "entrypoint" mode.
func IsEntrypoint() bool {
	return filepath.Dir(os.Args[0]) == EntrypointPrefix
}

func makeHandler() http.Handler {
	if !IsEntrypoint() {
		log.Fatal("makeHandler called on non-entrypoint: ", os.Args[0])
	}
	modeName := filepath.Base(os.Args[0])
	mode, ok := modes[modeName]
	if !ok {
		log.Fatal("Unsupported mode: ", modeName)
	}
	return mode()
}

// Entrypoint start up the queue in entrypoint mode.
// This should only be called when IsEntrypoint is true.
func Entrypoint(ctx context.Context) {
	h := &handlers.Drainer{Inner: makeHandler()}
	server := http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: h,
	}
	go server.ListenAndServe()

	// When the context is cancelled, start to drain.
	<-ctx.Done()
	h.Drain()
	server.Shutdown(context.Background())
}
