// Copyright 2016-2019 Flato Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prometheus

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bcds/go-hpc-dagbft/protocol/observer"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
)

func Test_Duplicate_Register(t *testing.T) {
	server, client, _ := prepare()

	prov := &Provider{Name: "test"}

	gOpts := observer.GaugeOpts{
		Name: "gauge",
	}
	g0, err := prov.NewGauge(gOpts)
	assert.Nil(t, err, "should register successfully")
	assert.NotNil(t, g0, "should register successfully")

	g0.Set(1)
	resp, err := client.Get(fmt.Sprintf("http://%s/metrics", server.Listener.Addr().String()))
	assert.Nil(t, err)
	bytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Contains(t, string(bytes), `test_gauge 1`)

	g1, err := prov.NewGauge(gOpts)
	assert.EqualError(t, err, prom.AlreadyRegisteredError{}.Error(), "should not register successfully because of duplicate name")
	assert.Nil(t, g1, "should not register successfully because of duplicate name")

	// unregister
	g0.Unregister()
	// register same metrics again
	g2, err := prov.NewGauge(gOpts)
	assert.Nil(t, err, "should not register successfully because of duplicate name")
	assert.NotNil(t, g2, "should not register successfully because of duplicate name")

	g2.Set(2)
	resp, err = client.Get(fmt.Sprintf("http://%s/metrics", server.Listener.Addr().String()))
	assert.Nil(t, err)
	bytes, err = io.ReadAll(resp.Body)
	assert.Nil(t, err)
	assert.Contains(t, string(bytes), `test_gauge 2`)
}

func Test_SubProvider(t *testing.T) {
	var p observer.MetricProvider
	p = &Provider{Name: "axiom_ledger"}
	prov := p.(*Provider)
	assert.Equal(t, "axiom_ledger", prov.Name)

	p = p.SubProvider("consensus").SubProvider("dagbft")

	prov = p.(*Provider)
	assert.Equal(t, "axiom_ledger_consensus_dagbft", prov.Name)
}

func Test_Counter(t *testing.T) {
	server, client, p := prepare()

	o := observer.CounterOpts{
		Name:       "tx",
		Help:       "test tx number",
		LabelNames: []string{"type"},
	}

	c, err := p.NewCounter(o)
	assert.Nil(t, err)
	c.With("type", "valid").Add(1)
	c.With("type", "invalid").Add(2)

	resp, err := client.Get(fmt.Sprintf("http://%s/metrics", server.Listener.Addr().String()))
	assert.Nil(t, err)
	bytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	assert.Contains(t, string(bytes), `# HELP dagbft_consensus_test_tx test tx number`)
	assert.Contains(t, string(bytes), `# TYPE dagbft_consensus_test_tx counter`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx{type="invalid"} 2`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx{type="valid"} 1`)
}

func Test_Gauge(t *testing.T) {
	server, client, p := prepare()

	o := observer.GaugeOpts{
		Name:       "tx",
		Help:       "test tx number",
		LabelNames: []string{"type"},
	}

	g, err := p.NewGauge(o)
	assert.Nil(t, err)
	g.With("type", "valid").Add(1)
	g.With("type", "invalid").Set(2)

	resp, err := client.Get(fmt.Sprintf("http://%s/metrics", server.Listener.Addr().String()))
	assert.Nil(t, err)
	bytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	assert.Contains(t, string(bytes), `# HELP dagbft_consensus_test_tx test tx number`)
	assert.Contains(t, string(bytes), `# TYPE dagbft_consensus_test_tx gauge`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx{type="invalid"} 2`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx{type="valid"} 1`)
}

func Test_Histogram(t *testing.T) {
	server, client, p := prepare()

	defBuckets := []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
	o := observer.HistogramOpts{
		Name:       "tx",
		Help:       "test tx number",
		LabelNames: []string{"type"},
		Buckets:    defBuckets,
	}

	h, err := p.NewHistogram(o)
	assert.Nil(t, err)
	for _, limit := range defBuckets {
		h.With("type", "valid").Observe(limit)
		h.With("type", "invalid").Observe(limit / 2)
	}

	resp, err := client.Get(fmt.Sprintf("http://%s/metrics", server.Listener.Addr().String()))
	assert.Nil(t, err)
	bytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	assert.Contains(t, string(bytes), `# HELP dagbft_consensus_test_tx test tx number`)
	assert.Contains(t, string(bytes), `# TYPE dagbft_consensus_test_tx histogram`)

	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_bucket{type="invalid",le="0.005"} 2`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_bucket{type="invalid",le="0.01"} 2`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_bucket{type="invalid",le="0.025"} 4`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_bucket{type="invalid",le="0.05"} 5`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_bucket{type="invalid",le="0.1"} 5`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_bucket{type="invalid",le="0.25"} 7`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_bucket{type="invalid",le="0.5"} 8`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_bucket{type="invalid",le="1"} 8`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_bucket{type="invalid",le="2.5"} 10`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_bucket{type="invalid",le="5"} 11`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_bucket{type="invalid",le="10"} 11`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_count{type="invalid"} 11`)

	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_bucket{type="valid",le="0.005"} 1`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_bucket{type="valid",le="0.01"} 2`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_bucket{type="valid",le="0.025"} 3`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_bucket{type="valid",le="0.05"} 4`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_bucket{type="valid",le="0.1"} 5`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_bucket{type="valid",le="0.25"} 6`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_bucket{type="valid",le="0.5"} 7`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_bucket{type="valid",le="1"} 8`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_bucket{type="valid",le="2.5"} 9`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_bucket{type="valid",le="5"} 10`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_bucket{type="valid",le="10"} 11`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_count{type="valid"} 11`)
}

func Test_Summary(t *testing.T) {
	server, client, p := prepare()

	defObj := map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}
	o := observer.SummaryOpts{
		Name:       "tx",
		Help:       "test tx number",
		LabelNames: []string{"type"},
		Objectives: defObj,
	}

	s, err := p.NewSummary(o)
	assert.Nil(t, err)
	s.With("type", "valid").Observe(1)
	s.With("type", "valid").Observe(3)
	s.With("type", "invalid").Observe(2)
	s.With("type", "invalid").Observe(4)

	resp, err := client.Get(fmt.Sprintf("http://%s/metrics", server.Listener.Addr().String()))
	assert.Nil(t, err)
	bytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	assert.Contains(t, string(bytes), `# HELP dagbft_consensus_test_tx test tx number`)
	assert.Contains(t, string(bytes), `# TYPE dagbft_consensus_test_tx summary`)

	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx{type="valid",quantile="0.5"} 1`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx{type="valid",quantile="0.9"} 3`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx{type="valid",quantile="0.99"} 3`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_sum{type="valid"} 4`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_count{type="valid"} 2`)

	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx{type="invalid",quantile="0.5"} 2`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx{type="invalid",quantile="0.9"} 4`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx{type="invalid",quantile="0.99"} 4`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_sum{type="invalid"} 6`)
	assert.Contains(t, string(bytes), `dagbft_consensus_test_tx_count{type="invalid"} 2`)
}

// Note: These tests can't be run in parallel because prometheus uses
// the global registry to manage observer.
func prepare() (*httptest.Server, *http.Client, observer.MetricProvider) {
	registry := prom.NewRegistry()
	prom.DefaultRegisterer = registry
	prom.DefaultGatherer = registry

	server := httptest.NewServer(promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	client := server.Client()

	var p observer.MetricProvider
	// construct sub-provider step by step
	p = &Provider{Name: "dagbft"}
	p = p.SubProvider("consensus")
	p = p.SubProvider("test")

	return server, client, p
}
