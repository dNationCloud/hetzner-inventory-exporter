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
	"fmt"
	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/collectors/version"
	"github.com/rkosegi/hetzner-inventory-exporter/internal"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	pv "github.com/prometheus/common/version"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"

	"github.com/prometheus/exporter-toolkit/web"
)

const (
	PROG_NAME = "hetzner_inventory_exporter"
)

var (
	webConfig     = webflag.AddFlags(kingpin.CommandLine, ":9112")
	telemetryPath = kingpin.Flag(
		"web.telemetry-path",
		"Path under which to expose metrics.",
	).Default("/metrics").String()
	disableDefaultMetrics = kingpin.Flag(
		"disable-default-metrics",
		"Exclude default metrics about the exporter itself (promhttp_*, process_*, go_*).",
	).Bool()
	configFile = kingpin.Flag(
		"config.file",
		"Path to YAML file with configuration",
	).Default("config.yaml").String()
)

func loadConfig(configFile string) (*internal.Config, error) {
	var cfg internal.Config
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func main() {
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(pv.Print(PROG_NAME))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promlog.New(promlogConfig)
	level.Info(logger).Log("msg", "Starting "+PROG_NAME, "version", pv.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", pv.BuildContext())
	level.Info(logger).Log("msg", "Loading config", "file", configFile)

	config, err := loadConfig(*configFile)

	if err != nil {
		level.Error(logger).Log("msg", "Error reading configuration", "err", err)
		os.Exit(1)
	}

	level.Info(logger).Log("msg", fmt.Sprintf("Got %d targets", len(config.Targets)))

	r := prometheus.NewRegistry()
	r.MustRegister(version.NewCollector(PROG_NAME))

	if err = r.Register(internal.New(config, logger)); err != nil {
		level.Error(logger).Log("msg", "Couldn't register "+PROG_NAME, "err", err)
		os.Exit(1)
	}
	handler := promhttp.HandlerFor(
		prometheus.Gatherers{r},
		promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
		},
	)

	if !*disableDefaultMetrics {
		r.MustRegister(collectors.NewGoCollector())
		r.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
		handler = promhttp.InstrumentMetricHandler(
			r, handler,
		)
	}

	landingPage, err := web.NewLandingPage(web.LandingConfig{
		Name:        strings.ReplaceAll(PROG_NAME, "_", " "),
		Description: "Prometheus Exporter for k8s API resources footprint",
		Version:     pv.Info(),
		Links: []web.LandingLinks{
			{
				Address: *telemetryPath,
				Text:    "Metrics",
			},
			{
				Address: "/health",
				Text:    "Health",
			},
		},
	})
	if err != nil {
		level.Error(logger).Log("msg", "Couldn't create landing page", "err", err)
		os.Exit(1)
	}

	http.Handle("/", landingPage)
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})
	http.Handle(*telemetryPath, handler)

	srv := &http.Server{
		ReadHeaderTimeout: 10 * time.Second,
	}
	if err := web.ListenAndServe(srv, webConfig, logger); err != nil {
		level.Error(logger).Log("msg", "Error starting server", "err", err)
		os.Exit(1)
	}
}
