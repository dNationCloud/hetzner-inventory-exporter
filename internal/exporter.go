// Copyright 2021 https://dnation.cloud
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"context"
	"sync"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "hetzner"
)

type Exporter struct {
	logger       log.Logger
	config       *Config
	scrapesSum   *prometheus.SummaryVec
	scrapeErrors *prometheus.CounterVec
	lastError    prometheus.Gauge
	scrapers     []scraper
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.lastError.Desc()
	e.scrapesSum.Describe(ch)
	e.scrapeErrors.Describe(ch)
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.scrape(context.TODO(), ch)
	e.scrapesSum.Collect(ch)
	ch <- e.lastError
	e.scrapeErrors.Collect(ch)
}

func (e *Exporter) scrapeTarget(ctx context.Context, ch chan<- prometheus.Metric, target string, c *hcloud.Client) {
	start := time.Now()
	defer func() {
		e.scrapesSum.WithLabelValues(target).Observe(time.Since(start).Seconds())
	}()
	var wg sync.WaitGroup
	defer wg.Wait()
	for _, s := range e.scrapers {
		s := s
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.Scrape(ctx, c, target, ch); err != nil {
				level.Error(e.logger).Log("msg", "lastError while scraping target", "target", target,
					"scraper", s.Name(), "err", err)
				e.scrapeErrors.WithLabelValues(s.Name()).Inc()
				e.lastError.Set(1)
			}
		}()
	}
}

func (e *Exporter) scrape(ctx context.Context, ch chan<- prometheus.Metric) {
	e.lastError.Set(0)

	var wg sync.WaitGroup
	defer wg.Wait()
	for _, target := range e.config.Targets {
		name := target.Name
		level.Debug(e.logger).Log("msg", "Scraping target", "target", name)
		client := hcloud.NewClient(hcloud.WithToken(target.ApiKey))
		wg.Add(1)
		go func() {
			defer wg.Done()
			e.scrapeTarget(ctx, ch, name, client)
		}()
	}
}

func New(config *Config, logger log.Logger) *Exporter {
	return &Exporter{
		logger: logger,
		config: config,
		scrapesSum: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Namespace: namespace,
			Subsystem: "exporter",
			Name:      "scrapes",
			Help:      "Scrapes summary on per-project basis.",
		}, []string{"project"}),
		scrapeErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "exporter",
			Name:      "scrape_errors_total",
			Help:      "Total number of times an error occurred scraping an inventory.",
		}, []string{"collector"}),
		lastError: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "exporter",
			Name:      "last_scrape_error",
			Help:      "Whether the last scrape of metrics from inventory resulted in an error (1 for error, 0 for success).",
		}),
		scrapers: []scraper{
			&imageScraper{},
			&serverScraper{},
			&volumeScraper{},
			&lbScraper{},
			&netScraper{},
		},
	}
}
