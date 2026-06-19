// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:generate mdatagen metadata.yaml

// Package nvidiareceiver implements an OpenTelemetry receiver that collects
// NVIDIA GPU metrics on a local host by invoking the nvidia-smi command.
package nvidiareceiver // import "github.com/pierre-prevoteau/otel-receiver-nvidia/nvidiareceiver"
