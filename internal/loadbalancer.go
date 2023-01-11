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
	lbCountDesc = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "inventory", "load_balancers"),
		"The number of load balancers in project.",
		[]string{"project"}, nil)
)

type lbScraper struct{}

func (l *lbScraper) Scrape(ctx context.Context, c *hcloud.Client, name string, ch chan<- prometheus.Metric) error {
	if lbs, _, err := c.LoadBalancer.List(ctx, hcloud.LoadBalancerListOpts{}); err != nil {
		return err
	} else {
		ch <- prometheus.MustNewConstMetric(lbCountDesc, prometheus.GaugeValue, float64(len(lbs)), name)
	}
	return nil
}

func (l *lbScraper) Name() string {
	return "load_balancer"
}
