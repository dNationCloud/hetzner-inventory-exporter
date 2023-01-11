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
	serverCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "inventory", "server_count"),
		"The number of servers in project.",
		[]string{"project"}, nil)
)

type serverScraper struct{}

func (s *serverScraper) Scrape(ctx context.Context, c *hcloud.Client, name string, ch chan<- prometheus.Metric) error {
	if servers, _, err := c.Server.List(ctx, hcloud.ServerListOpts{}); err != nil {
		return err
	} else {
		ch <- prometheus.MustNewConstMetric(serverCountDesc, prometheus.GaugeValue, float64(len(servers)), name)
	}
	return nil
}

func (s *serverScraper) Name() string {
	return "server"
}
