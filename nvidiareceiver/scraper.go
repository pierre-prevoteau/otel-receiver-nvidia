// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package nvidiareceiver // import "github.com/pierre-prevoteau/otel-receiver-nvidia/nvidiareceiver"

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/scraper/scrapererror"

	"github.com/pierre-prevoteau/otel-receiver-nvidia/nvidiareceiver/internal/metadata"
)

// queryGPUFields is the ordered list of nvidia-smi --query-gpu fields requested.
// The order must stay in sync with the indexes used in recordGPU.
const queryGPUFields = "index,uuid,name,utilization.gpu,memory.total,memory.used,memory.free"

const expectedFieldCount = 7

// mibToBytes converts the mebibytes that nvidia-smi reports memory in to bytes.
const mibToBytes int64 = 1024 * 1024

type nvidiaScraper struct {
	cfg       *Config
	telemetry component.TelemetrySettings
	mb        *metadata.MetricsBuilder

	// runSMI executes nvidia-smi and returns its raw stdout. It is a field so
	// tests can inject canned output without a real GPU being present.
	runSMI func(ctx context.Context) ([]byte, error)
}

func newScraper(cfg *Config, settings receiver.Settings) *nvidiaScraper {
	s := &nvidiaScraper{
		cfg:       cfg,
		telemetry: settings.TelemetrySettings,
		mb:        metadata.NewMetricsBuilder(cfg.MetricsBuilderConfig, settings),
	}
	s.runSMI = s.execNvidiaSMI
	return s
}

func (*nvidiaScraper) start(context.Context, component.Host) error {
	return nil
}

func (*nvidiaScraper) shutdown(context.Context) error {
	return nil
}

func (s *nvidiaScraper) execNvidiaSMI(ctx context.Context) ([]byte, error) {
	cmd := exec.CommandContext(ctx, s.cfg.BinaryPath,
		"--query-gpu="+queryGPUFields,
		"--format=csv,noheader,nounits",
	)

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed running %q: %w", s.cfg.BinaryPath, err)
	}
	return out, nil
}

func (s *nvidiaScraper) scrape(ctx context.Context) (pmetric.Metrics, error) {
	out, err := s.runSMI(ctx)
	if err != nil {
		return pmetric.NewMetrics(), err
	}

	reader := csv.NewReader(bytes.NewReader(out))
	reader.TrimLeadingSpace = true
	// GPU model names never contain commas, but be lenient about field counts
	// so a single malformed row does not abort the whole scrape.
	reader.FieldsPerRecord = -1

	rows, err := reader.ReadAll()
	if err != nil {
		return pmetric.NewMetrics(), fmt.Errorf("failed parsing nvidia-smi csv output: %w", err)
	}

	now := pcommon.NewTimestampFromTime(time.Now())
	var errs scrapererror.ScrapeErrors
	for _, row := range rows {
		s.recordGPU(now, row, &errs)
	}

	return s.mb.Emit(), errs.Combine()
}

func (s *nvidiaScraper) recordGPU(now pcommon.Timestamp, row []string, errs *scrapererror.ScrapeErrors) {
	if len(row) < expectedFieldCount {
		errs.AddPartial(1, fmt.Errorf("unexpected nvidia-smi row with %d fields (want %d): %q",
			len(row), expectedFieldCount, strings.Join(row, ",")))
		return
	}

	index, uuid, name := row[0], row[1], row[2]

	rb := s.mb.NewResourceBuilder()
	rb.SetNvidiaGpuIndex(index)
	rb.SetNvidiaGpuUUID(uuid)
	rb.SetNvidiaGpuName(name)

	if v, err := parseFloat(row[3]); err != nil {
		errs.AddPartial(1, fmt.Errorf("gpu %s: invalid utilization.gpu %q: %w", index, row[3], err))
	} else {
		s.mb.RecordNvidiaGpuUtilizationDataPoint(now, v)
	}

	totalMiB, totalErr := parseInt(row[4])
	if totalErr != nil {
		errs.AddPartial(1, fmt.Errorf("gpu %s: invalid memory.total %q: %w", index, row[4], totalErr))
	} else {
		s.mb.RecordNvidiaGpuMemoryTotalDataPoint(now, totalMiB*mibToBytes)
	}

	usedMiB, usedErr := parseInt(row[5])
	if usedErr != nil {
		errs.AddPartial(1, fmt.Errorf("gpu %s: invalid memory.used %q: %w", index, row[5], usedErr))
	} else {
		s.mb.RecordNvidiaGpuMemoryUsedDataPoint(now, usedMiB*mibToBytes)
	}

	if freeMiB, err := parseInt(row[6]); err != nil {
		errs.AddPartial(1, fmt.Errorf("gpu %s: invalid memory.free %q: %w", index, row[6], err))
	} else {
		s.mb.RecordNvidiaGpuMemoryFreeDataPoint(now, freeMiB*mibToBytes)
	}

	// Memory utilization is derived from used / total and only emitted when both
	// values parsed and total is positive (avoids a divide-by-zero).
	if totalErr == nil && usedErr == nil && totalMiB > 0 {
		s.mb.RecordNvidiaGpuMemoryUtilizationDataPoint(now, float64(usedMiB)/float64(totalMiB)*100)
	}

	s.mb.EmitForResource(metadata.WithResource(rb.Emit()))
}

func parseInt(s string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(s), 10, 64)
}

func parseFloat(s string) (float64, error) {
	return strconv.ParseFloat(strings.TrimSpace(s), 64)
}
