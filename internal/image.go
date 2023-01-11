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

type imageScraper struct{}

var (
	imageCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "inventory", "image_count"),
		"The number of images/snapshots in project.",
		[]string{"project"}, nil,
	)
	imageSizeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "inventory", "image_size"),
		"The summary size of images/snapshots in project (GiB).",
		[]string{"project"}, nil,
	)
)

func (i *imageScraper) Scrape(ctx context.Context, c *hcloud.Client, name string, ch chan<- prometheus.Metric) error {
	if images, _, err := c.Image.List(ctx, hcloud.ImageListOpts{
		Type: []hcloud.ImageType{hcloud.ImageTypeSnapshot, hcloud.ImageTypeBackup},
	}); err != nil {
		return err
	} else {
		ch <- prometheus.MustNewConstMetric(imageCountDesc, prometheus.GaugeValue, float64(len(images)), name)
		var size float32
		for _, image := range images {
			size += image.ImageSize
		}
		ch <- prometheus.MustNewConstMetric(imageSizeDesc, prometheus.GaugeValue, float64(size), name)
		return nil
	}
}

func (i *imageScraper) Name() string {
	return "image"
}
