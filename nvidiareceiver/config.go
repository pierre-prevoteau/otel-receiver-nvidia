// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package nvidiareceiver // import "github.com/pierre-prevoteau/otel-receiver-nvidia/nvidiareceiver"

import (
	"errors"

	"go.opentelemetry.io/collector/scraper/scraperhelper"

	"github.com/pierre-prevoteau/otel-receiver-nvidia/nvidiareceiver/internal/metadata"
)

var errEmptyBinaryPath = errors.New(`"binary_path" must not be empty`)

// Config defines configuration for the NVIDIA GPU receiver.
type Config struct {
	scraperhelper.ControllerConfig `mapstructure:",squash"`
	metadata.MetricsBuilderConfig  `mapstructure:",squash"`

	// BinaryPath is the path to the nvidia-smi executable. When set to a bare
	// file name (the default, "nvidia-smi") it is resolved using the PATH.
	BinaryPath string `mapstructure:"binary_path"`
}

// Validate checks that the receiver configuration is valid.
func (c Config) Validate() error {
	if c.BinaryPath == "" {
		return errEmptyBinaryPath
	}
	return nil
}
