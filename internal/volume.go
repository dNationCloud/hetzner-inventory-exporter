// Copyright 2023 https://dnation.cloud
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

	"github.com/hetznercloud/hcloud-go/hcloud"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	volumeCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "inventory", "volume_count"),
		"The number of volumes in project.",
		[]string{"project"}, nil)
	volumeSizeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "inventory", "volume_size"),
		"The number of volumes in project.",
		[]string{"project"}, nil)
)

type volumeScraper struct{}

func (v *volumeScraper) Scrape(ctx context.Context, c *hcloud.Client, name string, ch chan<- prometheus.Metric) error {
	if volumes, _, err := c.Volume.List(ctx, hcloud.VolumeListOpts{}); err != nil {
		return err
	} else {
		ch <- prometheus.MustNewConstMetric(volumeCountDesc, prometheus.GaugeValue, float64(len(volumes)), name)
		var size int
		for _, volume := range volumes {
			size += volume.Size
		}
		ch <- prometheus.MustNewConstMetric(volumeSizeDesc, prometheus.GaugeValue, float64(size), name)
	}
	return nil
}

func (v *volumeScraper) Name() string {
	return "volume"
}
