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
	netCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "inventory", "network_count"),
		"The number of networks in project.",
		[]string{"project"}, nil)
)

type netScraper struct{}

func (n *netScraper) Scrape(ctx context.Context, c *hcloud.Client, name string, ch chan<- prometheus.Metric) error {
	if nets, _, err := c.Network.List(ctx, hcloud.NetworkListOpts{}); err != nil {
		return err
	} else {
		ch <- prometheus.MustNewConstMetric(netCountDesc, prometheus.GaugeValue, float64(len(nets)), name)
	}
	return nil
}

func (n *netScraper) Name() string {
	return "network"
}
