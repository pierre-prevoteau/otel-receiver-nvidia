// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package nvidiareceiver // import "github.com/pierre-prevoteau/otel-receiver-nvidia/nvidiareceiver"

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/scraper"
	"go.opentelemetry.io/collector/scraper/scraperhelper"

	"github.com/pierre-prevoteau/otel-receiver-nvidia/nvidiareceiver/internal/metadata"
)

// defaultBinaryPath is resolved against the PATH at scrape time.
const defaultBinaryPath = "nvidia-smi"

var errConfigNotValid = errors.New("config is not valid for the nvidia receiver")

// NewFactory creates a factory for the NVIDIA GPU receiver.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		metadata.Type,
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, metadata.MetricsStability))
}

func createDefaultConfig() component.Config {
	cfg := scraperhelper.NewDefaultControllerConfig()
	cfg.CollectionInterval = 60 * time.Second

	return &Config{
		ControllerConfig:     cfg,
		MetricsBuilderConfig: metadata.NewDefaultMetricsBuilderConfig(),
		BinaryPath:           defaultBinaryPath,
	}
}

func createMetricsReceiver(_ context.Context, params receiver.Settings, rConf component.Config, consumer consumer.Metrics) (receiver.Metrics, error) {
	cfg, ok := rConf.(*Config)
	if !ok {
		return nil, errConfigNotValid
	}

	ns := newScraper(cfg, params)
	s, err := scraper.NewMetrics(ns.scrape, scraper.WithStart(ns.start), scraper.WithShutdown(ns.shutdown))
	if err != nil {
		return nil, err
	}

	return scraperhelper.NewMetricsController(
		&cfg.ControllerConfig,
		params,
		consumer,
		scraperhelper.AddMetricsScraper(metadata.Type, s),
	)
}
