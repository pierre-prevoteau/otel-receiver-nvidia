// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package nvidiareceiver

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/golden"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatatest/pmetrictest"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/receiver/receivertest"

	"github.com/pierre-prevoteau/otel-receiver-nvidia/nvidiareceiver/internal/metadata"
)

// writeGolden controls whether the expected golden file is (re)generated from
// the current scrape output. Run `WRITE_GOLDEN=true go test ./...` to refresh
// the testdata after intentionally changing the emitted metrics.
var writeGolden = os.Getenv("WRITE_GOLDEN") == "true"

// sampleOutput mimics `nvidia-smi --query-gpu=index,uuid,name,utilization.gpu,
// memory.total,memory.used,memory.free --format=csv,noheader,nounits` for a host
// with two GPUs (memory values are in MiB).
const sampleOutput = `0, GPU-aaaaaaaa-1111-2222-3333-444444444444, NVIDIA A100-SXM4-40GB, 42, 40960, 8192, 32768
1, GPU-bbbbbbbb-5555-6666-7777-888888888888, NVIDIA A100-SXM4-40GB, 0, 40960, 0, 40960
`

func newTestScraper(t *testing.T, output string, runErr error) *nvidiaScraper {
	t.Helper()
	cfg := createDefaultConfig().(*Config)
	s := newScraper(cfg, receivertest.NewNopSettings(metadata.Type))
	s.runSMI = func(context.Context) ([]byte, error) {
		if runErr != nil {
			return nil, runErr
		}
		return []byte(output), nil
	}
	return s
}

func TestScrape(t *testing.T) {
	s := newTestScraper(t, sampleOutput, nil)

	actual, err := s.scrape(context.Background())
	require.NoError(t, err)

	expectedFile := filepath.Join("testdata", "expected_metrics.yaml")
	if writeGolden {
		require.NoError(t, golden.WriteMetrics(t, expectedFile, actual))
	}

	expected, err := golden.ReadMetrics(expectedFile)
	require.NoError(t, err)

	require.NoError(t, pmetrictest.CompareMetrics(expected, actual,
		pmetrictest.IgnoreStartTimestamp(),
		pmetrictest.IgnoreTimestamp(),
		pmetrictest.IgnoreResourceMetricsOrder(),
		pmetrictest.IgnoreMetricsOrder(),
	))
}

func TestScrapeCommandError(t *testing.T) {
	wantErr := errors.New("nvidia-smi: command not found")
	s := newTestScraper(t, "", wantErr)

	_, err := s.scrape(context.Background())
	require.ErrorIs(t, err, wantErr)
}

func TestScrapePartialRow(t *testing.T) {
	// utilization.gpu is not a number; the memory metrics for the GPU should
	// still be emitted while a partial scrape error is reported.
	s := newTestScraper(t, "0, GPU-x, NVIDIA T4, N/A, 16384, 1024, 15360\n", nil)

	m, err := s.scrape(context.Background())
	require.Error(t, err)
	require.Equal(t, 1, m.ResourceMetrics().Len())
}

func TestScrapeMalformedRowSkipped(t *testing.T) {
	// First row has too few fields and is skipped with a partial error; the
	// second row is valid and produces one resource.
	s := newTestScraper(t, "garbage,row\n2, GPU-y, NVIDIA L4, 10, 24576, 2048, 22528\n", nil)

	m, err := s.scrape(context.Background())
	require.Error(t, err)
	require.Equal(t, 1, m.ResourceMetrics().Len())
}
