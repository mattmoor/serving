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

package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

const tpl = `
<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
  </head>
  <body>
    <table width="100%">
      <tr>
        <td width="15%" colspan=3><b><u>Scenario</u></b></td>
        <td width="42%"><b><u>Reliability</u></b></td>
        <td width="43%"><b><u>Performance</u></b></td>
      </tr>
      {{range .Scenarios}}
        <tr>
          <td colspan=3>{{ .Name }}</td>
          <td></td>
          <td></td>
        </tr>
        {{range .Benchmarks}}
          <tr>
            <td></td>
            <td colspan=2>{{ .Name }}</td>
            <td></td>
            <td></td>
          </tr>
          {{range .Configurations}}
            <tr>
              <td width="1%"></td>
              <td width="1%"></td>
              <td>{{ .Name }}</td>
              <td>{{ .Reliability }}</td>
              <td>{{ .Performance }}</td>
            </tr>
            <tr>
              <td colspan=3></td>
              {{ if ne .ReliabilityURL.String "" }}
                <td><iframe width="100%" height=200 src="{{ .ReliabilityURL.String }}"></iframe></td>
              {{ else }}
                <td></td>
              {{ end }}
              {{ if ne .PerformanceURL.String "" }}
                <td><iframe width="100%" height=200 src="{{ .PerformanceURL.String }}"></iframe></td>
              {{ else }}
                <td></td>
              {{ end }}
            </tr>
          {{else}}
            <tr>
              <td colspan=5><i>no configurations</i></td>
            </tr>
          {{end}}
        {{else}}
          <tr>
            <td colspan=5><i>no benchmarks</i></td>
          </tr>
        {{end}}
      {{else}}
        <tr>
          <td colspan=5><i>no scenarios</i></td>
        </tr>
      {{end}}
    </table>
  </body>
</html>`

type Dashboard struct {
	Title     string
	Scenarios []Scenario
}

type Scenario struct {
	Name       string
	Benchmarks []Benchmark
}

type Benchmark struct {
	Name           string
	Configurations []Configuration
}

type Configuration struct {
	Name           string
	Reliability    string
	ReliabilityURL url.URL
	Performance    string
	PerformanceURL url.URL
}

func main() {
	t, err := template.New("webpage").Parse(tpl)
	if err != nil {
		log.Fatal(err)
	}

	duration := fmt.Sprintf("%d", int((24 * time.Hour).Seconds()))

	// TODO(mattmoor): Build this programmatically, this is too regular...
	data := Dashboard{
		Title: "knative/serving dashboard",
		Scenarios: []Scenario{{
			Name: "Control Plane: Deployment Latency",
			Benchmarks: []Benchmark{{
				Name: "Service per Xs for Ym",
				Configurations: []Configuration{{
					Name:        "X=5s, Y=35m",
					Reliability: "99.9%",
					ReliabilityURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5143375149793280"},
							"tseconds":      []string{duration},
							"stacked":       []string{"1"},
							"embed":         []string{"1"},

							// The tags to show
							"tag": []string{"frequency=5s", "duration=35m0s"},

							// The metrics
							"~dl":         []string{"count"},
							"error-count": []string{"1"},
						}.Encode(),
					},
					Performance: "25s @ 95P",
					PerformanceURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5143375149793280"},
							"tseconds":      []string{duration},
							"embed":         []string{"1"},

							// The tags to show
							"tag": []string{"frequency=5s", "duration=35m0s"},

							// The metrics
							"~dl": []string{"p95000"},
						}.Encode(),
					},
				}},
			}},
		}, {
			Name: "Control Plane: # Services",
			Benchmarks: []Benchmark{{
				Name: "Service per Xs for Ym",
				Configurations: []Configuration{{
					Name:        "X=5s, Y=35m",
					Reliability: "(see above)",
					Performance: "410 @ count",
					PerformanceURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5143375149793280"},
							"tseconds":      []string{duration},
							"embed":         []string{"1"},

							// The tags to show
							"tag": []string{"frequency=5s", "duration=35m0s"},

							// The metrics
							"~dl": []string{"count"},
						}.Encode(),
					},
				}},
			}},
		}, {
			Name: "Data Plane: Cold Start Latency",
			Benchmarks: []Benchmark{{
				Name: "0 -> 1k -> 2k -> 3k",
				Configurations: []Configuration{{
					Name:        "TBC=0",
					Reliability: "(see overload)",
					Performance: "10s @ max",
					PerformanceURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5352009922248704"},
							"tseconds":      []string{duration},
							"embed":         []string{"1"},

							// The tags to show
							"tag": []string{"tbc=zero"},

							// The metrics
							"~l": []string{"max"},
						}.Encode(),
					},
				}, {
					Name:        "TBC=200 (default)",
					Reliability: "(see overload)",
					Performance: "10s @ max",
					PerformanceURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5352009922248704"},
							"tseconds":      []string{duration},
							"embed":         []string{"1"},

							// The tags to show
							"tag": []string{"tbc=200"},

							// The metrics
							"~l": []string{"max"},
						}.Encode(),
					},
				}, {
					Name:        "TBC=-1",
					Reliability: "(see overload)",
					Performance: "10s @ max",
					PerformanceURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5352009922248704"},
							"tseconds":      []string{duration},
							"embed":         []string{"1"},

							// The tags to show
							"tag": []string{"tbc=always"},

							// The metrics
							"~l": []string{"max"},
						}.Encode(),
					},
				}},
			}},
		}, {
			Name: "Data Plane: Overload",
			Benchmarks: []Benchmark{{
				Name: "0 -> 1k -> 2k -> 3k",
				Configurations: []Configuration{{
					Name:        "TBC=0",
					Reliability: "99.9%",
					ReliabilityURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5352009922248704"},
							"tseconds":      []string{duration},
							"stacked":       []string{"1"},
							"embed":         []string{"1"},

							// The tags to show
							"tag": []string{"tbc=zero"},

							// The metrics
							"~l":          []string{"count"},
							"error-count": []string{"1"},
						}.Encode(),
					},
					Performance: "15ms @ 95P",
					PerformanceURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5352009922248704"},
							"tseconds":      []string{duration},
							"embed":         []string{"1"},

							// The tags to show
							"tag": []string{"tbc=zero"},

							// The metrics
							"~l": []string{"p95000"},
						}.Encode(),
					},
				}, {
					Name:        "TBC=200 (default)",
					Reliability: "99.9%",
					ReliabilityURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5352009922248704"},
							"tseconds":      []string{duration},
							"stacked":       []string{"1"},
							"embed":         []string{"1"},

							// The tags to show
							"tag": []string{"tbc=200"},

							// The metrics
							"~l":          []string{"count"},
							"error-count": []string{"1"},
						}.Encode(),
					},
					Performance: "15ms @ 95P",
					PerformanceURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5352009922248704"},
							"tseconds":      []string{duration},
							"embed":         []string{"1"},

							// The tags to show
							"tag": []string{"tbc=200"},

							// The metrics
							"~l": []string{"p95000"},
						}.Encode(),
					},
				}, {
					Name:        "TBC=-1",
					Reliability: "99.9%",
					ReliabilityURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5352009922248704"},
							"tseconds":      []string{duration},
							"stacked":       []string{"1"},
							"embed":         []string{"1"},

							// The tags to show
							"tag": []string{"tbc=always"},

							// The metrics
							"~l":          []string{"count"},
							"error-count": []string{"1"},
						}.Encode(),
					},
					Performance: "15ms @ 95P",
					PerformanceURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5352009922248704"},
							"tseconds":      []string{duration},
							"embed":         []string{"1"},

							// The tags to show
							"tag": []string{"tbc=always"},

							// The metrics
							"~l": []string{"p95000"},
						}.Encode(),
					},
				}},
			}},
		}, {
			Name: "Data Plane: Steady State",
			Benchmarks: []Benchmark{{
				Name: "queue",
				Configurations: []Configuration{{
					Name:        "containerConcurrency: 0 (default)",
					Reliability: "99.99%",
					ReliabilityURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5142965274017792"},
							"tseconds":      []string{duration},
							"embed":         []string{"1"},

							// The metrics
							"~ke": []string{"p95000"},
							"~ie": []string{"p95000"},
							"~qe": []string{"p95000"},
						}.Encode(),
					},
					Performance: "10ms @ 95P",
					PerformanceURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5142965274017792"},
							"tseconds":      []string{duration},
							"embed":         []string{"1"},

							// The metrics
							"~kd": []string{"p95000"},
							"~id": []string{"p95000"},
							"~qp": []string{"p95000"},
						}.Encode(),
					},
				}, {
					Name:        "containerConcurrency: 100",
					Reliability: "99.99%",
					ReliabilityURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5142965274017792"},
							"tseconds":      []string{duration},
							"embed":         []string{"1"},

							// The metrics
							"~ke": []string{"p95000"},
							"~ie": []string{"p95000"},
							"~re": []string{"p95000"},
						}.Encode(),
					},
					Performance: "10ms @ 95P",
					PerformanceURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5142965274017792"},
							"tseconds":      []string{duration},
							"embed":         []string{"1"},

							// The metrics
							"~kd": []string{"p95000"},
							"~id": []string{"p95000"},
							"~qc": []string{"p95000"},
						}.Encode(),
					},
				}, {
					Name:        "containerConcurrency: 10",
					Reliability: "99.99%",
					ReliabilityURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5142965274017792"},
							"tseconds":      []string{duration},
							"embed":         []string{"1"},

							// The metrics
							"~ke":  []string{"p95000"},
							"~ie":  []string{"p95000"},
							"~ret": []string{"p95000"},
						}.Encode(),
					},
					Performance: "TBD",
					PerformanceURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5142965274017792"},
							"tseconds":      []string{duration},
							"embed":         []string{"1"},

							// The metrics
							"~kd":  []string{"p95000"},
							"~id":  []string{"p95000"},
							"~qct": []string{"p95000"},
						}.Encode(),
					},
				}, {
					Name:        "containerConcurrency: 1",
					Reliability: "99.99%",
					ReliabilityURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5142965274017792"},
							"tseconds":      []string{duration},
							"embed":         []string{"1"},

							// The metrics
							"~ke":  []string{"p95000"},
							"~ie":  []string{"p95000"},
							"~re1": []string{"p95000"},
						}.Encode(),
					},
					Performance: "TBD",
					PerformanceURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5142965274017792"},
							"tseconds":      []string{duration},
							"embed":         []string{"1"},

							// The metrics
							"~kd":  []string{"p95000"},
							"~id":  []string{"p95000"},
							"~qc1": []string{"p95000"},
						}.Encode(),
					},
				}},
			}, {
				Name: "activator",
				Configurations: []Configuration{{
					Name:        "containerConcurrency: 0 (default)",
					Reliability: "99.99%",
					ReliabilityURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5142965274017792"},
							"tseconds":      []string{duration},
							"embed":         []string{"1"},

							// The metrics
							"~ke": []string{"p95000"},
							"~ie": []string{"p95000"},
							"~qe": []string{"p95000"},
							"~ae": []string{"p95000"},
						}.Encode(),
					},
					Performance: "10ms @ 95P",
					PerformanceURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5142965274017792"},
							"tseconds":      []string{duration},
							"embed":         []string{"1"},

							// The metrics
							"~kd": []string{"p95000"},
							"~id": []string{"p95000"},
							"~qp": []string{"p95000"},
							"~a":  []string{"p95000"},
						}.Encode(),
					},
				}, {
					Name:        "containerConcurrency: 100",
					Reliability: "99.99%",
					ReliabilityURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5142965274017792"},
							"tseconds":      []string{duration},
							"embed":         []string{"1"},

							// The metrics
							"~ke": []string{"p95000"},
							"~ie": []string{"p95000"},
							"~re": []string{"p95000"},
							"~be": []string{"p95000"},
						}.Encode(),
					},
					Performance: "10ms @ 95P",
					PerformanceURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5142965274017792"},
							"tseconds":      []string{duration},
							"embed":         []string{"1"},

							// The metrics
							"~kd": []string{"p95000"},
							"~id": []string{"p95000"},
							"~qc": []string{"p95000"},
							"~ac": []string{"p95000"},
						}.Encode(),
					},
				}, {
					Name:        "containerConcurrency: 10",
					Reliability: "99.99%",
					ReliabilityURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5142965274017792"},
							"tseconds":      []string{duration},
							"embed":         []string{"1"},

							// The metrics
							"~ke":  []string{"p95000"},
							"~ie":  []string{"p95000"},
							"~ret": []string{"p95000"},
							"~bet": []string{"p95000"},
						}.Encode(),
					},
					Performance: "TBD",
					PerformanceURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5142965274017792"},
							"tseconds":      []string{duration},
							"embed":         []string{"1"},

							// The metrics
							"~kd":  []string{"p95000"},
							"~id":  []string{"p95000"},
							"~qct": []string{"p95000"},
							"~act": []string{"p95000"},
						}.Encode(),
					},
				}, {
					Name:        "containerConcurrency: 1",
					Reliability: "99.99%",
					ReliabilityURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5142965274017792"},
							"tseconds":      []string{duration},
							"embed":         []string{"1"},

							// The metrics
							"~ke":  []string{"p95000"},
							"~ie":  []string{"p95000"},
							"~re1": []string{"p95000"},
							"~be1": []string{"p95000"},
						}.Encode(),
					},
					Performance: "TBD",
					PerformanceURL: url.URL{
						Scheme: "https",
						Host:   "mako.dev",
						Path:   "/benchmark",
						RawQuery: url.Values{
							"benchmark_key": []string{"5142965274017792"},
							"tseconds":      []string{duration},
							"embed":         []string{"1"},

							// The metrics
							"~kd":  []string{"p95000"},
							"~id":  []string{"p95000"},
							"~qc1": []string{"p95000"},
							"~ac1": []string{"p95000"},
						}.Encode(),
					},
				}},
			}},
		}},
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := t.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
