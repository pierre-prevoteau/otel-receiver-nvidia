// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package nvidiareceiver

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"

	"github.com/pierre-prevoteau/otel-receiver-nvidia/nvidiareceiver/internal/metadata"
)

func TestLoadConfig(t *testing.T) {
	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)

	customExpected := createDefaultConfig().(*Config)
	customExpected.CollectionInterval = 30 * time.Second
	customExpected.BinaryPath = "/opt/nvidia/bin/nvidia-smi"

	tests := []struct {
		id       component.ID
		expected component.Config
	}{
		{
			id:       component.NewIDWithName(metadata.Type, ""),
			expected: createDefaultConfig(),
		},
		{
			id:       component.NewIDWithName(metadata.Type, "custom"),
			expected: customExpected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			cfg := createDefaultConfig()
			sub, err := cm.Sub(tt.id.String())
			require.NoError(t, err)
			require.NoError(t, sub.Unmarshal(cfg))

			assert.Equal(t, tt.expected, cfg)
		})
	}
}

func TestValidate(t *testing.T) {
	cfg := createDefaultConfig().(*Config)
	require.NoError(t, cfg.Validate())

	cfg.BinaryPath = ""
	require.ErrorIs(t, cfg.Validate(), errEmptyBinaryPath)
}
