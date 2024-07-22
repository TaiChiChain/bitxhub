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

package disabled

import (
	"github.com/bcds/go-hpc-dagbft/protocol/observer"
)

// Provider is the NIL implementation instance.
type Provider struct{}

// SubProvider creates a new instance of Provider with given suffix.
func (p *Provider) SubProvider(string) observer.MetricProvider {
	return p
}

// SetNamespace set given Namespace to a provider.
func (p *Provider) SetNamespace(string) {
	return
}

// NewCounter constructs a Counter.
func (p *Provider) NewCounter(observer.CounterOpts) (observer.Counter, error) { return &Counter{}, nil }

// NewGauge constructs a Gauge.
func (p *Provider) NewGauge(observer.GaugeOpts) (observer.Gauge, error) { return &Gauge{}, nil }

// NewHistogram constructs a Histogram.
func (p *Provider) NewHistogram(observer.HistogramOpts) (observer.Histogram, error) {
	return &Histogram{}, nil
}

// NewSummary constructs a Summary.
func (p *Provider) NewSummary(observer.SummaryOpts) (observer.Summary, error) { return &Summary{}, nil }

// Counter is an empty Counter instance.
type Counter struct{}

// Add implements the Counter interface.
func (c *Counter) Add(float64) {}

// With implements the Counter interface.
func (c *Counter) With(...string) observer.Counter {
	return c
}

// Unregister implements the Counter interface.
func (c *Counter) Unregister() {}

// Gauge is an empty Gauge instance.
type Gauge struct{}

// Add implements the Gauge interface.
func (g *Gauge) Add(float64) {}

// Set implements the Gauge interface.
func (g *Gauge) Set(float64) {}

// With implements the Gauge interface.
func (g *Gauge) With(...string) observer.Gauge {
	return g
}

// Unregister implements the Gauge interface.
func (g *Gauge) Unregister() {}

// Histogram is an empty Histogram instance.
type Histogram struct{}

// Observe implements the Histogram interface.
func (h *Histogram) Observe(float64) {}

// With implements the Histogram interface.
func (h *Histogram) With(...string) observer.Histogram {
	return h
}

// Unregister implements the Histogram interface.
func (h *Histogram) Unregister() {}

// Summary is an empty Summary instance.
type Summary struct{}

// Observe implements the Summary interface.
func (s *Summary) Observe(float64) {}

// With implements the Summary interface.
func (s *Summary) With(...string) observer.Summary {
	return s
}

// Unregister implements the Summary interface.
func (s *Summary) Unregister() {}
