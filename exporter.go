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

package main

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "hetzner"
)

var (
	serverCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "inventory", "server_count"),
		"The number of servers in project.",
		[]string{"project"}, nil)
	volumeCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "inventory", "volume_count"),
		"The number of volumes in project.",
		[]string{"project"}, nil)
	volumeSizeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "inventory", "volume_size"),
		"The number of volumes in project.",
		[]string{"project"}, nil)
	lbCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "inventory", "load_balancers"),
		"The number of load balancers in project.",
		[]string{"project"}, nil)
	netCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "inventory", "network_count"),
		"The number of networks in project.",
		[]string{"project"}, nil)
	imageCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "inventory", "image_count"),
		"The number of images/snapshots in project.",
		[]string{"project"}, nil)
	imageSizeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "inventory", "image_size"),
		"The summary size of images/snapshots in project.",
		[]string{"project"}, nil)
)

type Exporter struct {
	ctx     context.Context
	logger  log.Logger
	config  *Config
	metrics ExporterMetrics
	//client  client.OwmClient
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.metrics.TotalScrapes.Desc()
	ch <- e.metrics.Error.Desc()
	e.metrics.ScrapeErrors.Describe(ch)
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.scrape(e.ctx, ch)

	ch <- e.metrics.TotalScrapes
	ch <- e.metrics.Error

	e.metrics.ScrapeErrors.Collect(ch)
}

func (e *Exporter) scrape(ctx context.Context, ch chan<- prometheus.Metric) {
	e.metrics.TotalScrapes.Inc()
	e.metrics.Error.Set(0)

	for _, target := range e.config.Targets {
		level.Debug(e.logger).Log("msg", "Processing target", "target", target.Name)

		client := hcloud.NewClient(hcloud.WithToken(target.ApiKey))

		servers, _, err := client.Server.List(context.Background(), hcloud.ServerListOpts{})
		if err != nil {
			level.Error(e.logger).Log("msg", "Error while fetching list of servers",
				"name", target.Name, "err", err)
			e.metrics.ScrapeErrors.WithLabelValues("collect.serverlist." + target.Name).Inc()
			e.metrics.Error.Set(1)
		} else {

			ch <- prometheus.MustNewConstMetric(serverCountDesc,
				prometheus.GaugeValue, float64(len(servers)), target.Name)
		}

		images, _, err := client.Image.List(context.Background(), hcloud.ImageListOpts{
			Type: []hcloud.ImageType{hcloud.ImageTypeSnapshot, hcloud.ImageTypeBackup},
		})
		if err != nil {
			level.Error(e.logger).Log("msg", "Error while fetching list of images",
				"name", target.Name, "err", err)
			e.metrics.ScrapeErrors.WithLabelValues("collect.imagelist." + target.Name).Inc()
			e.metrics.Error.Set(1)
		} else {
			ch <- prometheus.MustNewConstMetric(imageCountDesc,
				prometheus.GaugeValue, float64(len(images)), target.Name)
			var size float32
			for _, image := range images {
				size += image.ImageSize
			}
			ch <- prometheus.MustNewConstMetric(imageSizeDesc,
				prometheus.GaugeValue, float64(size), target.Name)

		}

		volumes, _, err := client.Volume.List(context.Background(), hcloud.VolumeListOpts{})
		if err != nil {
			level.Error(e.logger).Log("msg", "Error while fetching list of volumes",
				"name", target.Name, "err", err)
			e.metrics.ScrapeErrors.WithLabelValues("collect.volumelist." + target.Name).Inc()
			e.metrics.Error.Set(1)
		} else {
			ch <- prometheus.MustNewConstMetric(volumeCountDesc,
				prometheus.GaugeValue, float64(len(volumes)), target.Name)
			var size int
			for _, volume := range volumes {
				size += volume.Size
			}
			ch <- prometheus.MustNewConstMetric(volumeSizeDesc,
				prometheus.GaugeValue, float64(size), target.Name)

		}

		lbs, _, err := client.LoadBalancer.List(context.Background(), hcloud.LoadBalancerListOpts{})
		if err != nil {
			level.Error(e.logger).Log("msg", "Error while fetching list of load balancers",
				"name", target.Name, "err", err)
			e.metrics.ScrapeErrors.WithLabelValues("collect.lblist." + target.Name).Inc()
			e.metrics.Error.Set(1)
		} else {
			ch <- prometheus.MustNewConstMetric(lbCountDesc,
				prometheus.GaugeValue, float64(len(lbs)), target.Name)
		}

		nets, _, err := client.Network.List(context.Background(), hcloud.NetworkListOpts{})
		if err != nil {
			level.Error(e.logger).Log("msg", "Error while fetching list of networks",
				"name", target.Name, "err", err)
			e.metrics.ScrapeErrors.WithLabelValues("collect.lblist." + target.Name).Inc()
			e.metrics.Error.Set(1)
		} else {
			ch <- prometheus.MustNewConstMetric(netCountDesc,
				prometheus.GaugeValue, float64(len(nets)), target.Name)
		}

	}
}

func newExporter(ctx context.Context, config *Config, logger log.Logger,
	exporterMetrics ExporterMetrics) *Exporter {
	return &Exporter{
		ctx:     ctx,
		logger:  logger,
		config:  config,
		metrics: exporterMetrics,
	}
}

func newExporterMetrics() ExporterMetrics {
	return ExporterMetrics{
		TotalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "exporter",
			Name:      "scrapes_total",
			Help:      "Total number of times inventory was scraped for metrics.",
		}),
		ScrapeErrors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "exporter",
			Name:      "scrape_errors_total",
			Help:      "Total number of times an error occurred scraping an inventory.",
		}, []string{"collector"}),
		Error: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "exporter",
			Name:      "last_scrape_error",
			Help:      "Whether the last scrape of metrics from inventory resulted in an error (1 for error, 0 for success).",
		}),
	}
}
