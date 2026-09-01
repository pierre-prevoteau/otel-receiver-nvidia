// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build !windows

package nvidiareceiver // import "github.com/pierre-prevoteau/otel-receiver-nvidia/nvidiareceiver"

import "os/exec"

// resolveBinaryPath leaves the configured path untouched: exec resolves bare
// file names against the PATH.
func resolveBinaryPath(configured string) (string, error) {
	return configured, nil
}

func hideConsoleWindow(*exec.Cmd) {}
