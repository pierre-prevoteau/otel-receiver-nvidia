// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build windows

package nvidiareceiver

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// isolateEnv empties every variable resolveBinaryPath consults so a test only
// sees the directories it sets up itself.
func isolateEnv(t *testing.T) {
	t.Helper()
	for _, env := range []string{"PATH", "SystemRoot", "ProgramW6432", "ProgramFiles", "ProgramFiles(x86)"} {
		t.Setenv(env, "")
	}
}

func TestResolveBinaryPathKeepsExplicitPath(t *testing.T) {
	isolateEnv(t)

	const explicit = `C:\opt\nvidia\nvidia-smi.exe`
	path, err := resolveBinaryPath(explicit)
	require.NoError(t, err)
	require.Equal(t, explicit, path)
}

func TestResolveBinaryPathFromPATH(t *testing.T) {
	isolateEnv(t)

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "nvidia-smi.exe"), nil, 0o600))
	t.Setenv("PATH", dir)

	path, err := resolveBinaryPath("nvidia-smi")
	require.NoError(t, err)
	require.Equal(t, filepath.Join(dir, "nvidia-smi.exe"), path)
}

// Drivers install nvidia-smi.exe into System32, and older ones into the NVSMI
// folder under Program Files; a service PATH does not always cover either.
func TestResolveBinaryPathFromInstallDir(t *testing.T) {
	tests := map[string]struct {
		env    string
		subdir []string
	}{
		"System32": {env: "SystemRoot", subdir: []string{"System32"}},
		"NVSMI":    {env: "ProgramFiles", subdir: []string{"NVIDIA Corporation", "NVSMI"}},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			isolateEnv(t)

			root := t.TempDir()
			dir := filepath.Join(append([]string{root}, test.subdir...)...)
			require.NoError(t, os.MkdirAll(dir, 0o750))
			require.NoError(t, os.WriteFile(filepath.Join(dir, "nvidia-smi.exe"), nil, 0o600))
			t.Setenv(test.env, root)

			path, err := resolveBinaryPath("nvidia-smi")
			require.NoError(t, err)
			require.Equal(t, filepath.Join(dir, "nvidia-smi.exe"), path)
		})
	}
}

func TestResolveBinaryPathNotFound(t *testing.T) {
	isolateEnv(t)
	t.Setenv("SystemRoot", t.TempDir())

	_, err := resolveBinaryPath("nvidia-smi")
	require.ErrorContains(t, err, "not found on PATH")
	require.ErrorContains(t, err, "System32")
}

func TestHideConsoleWindow(t *testing.T) {
	cmd := exec.Command("nvidia-smi")
	hideConsoleWindow(cmd)

	require.NotNil(t, cmd.SysProcAttr)
	require.True(t, cmd.SysProcAttr.HideWindow)
	require.Equal(t, uint32(createNoWindow), cmd.SysProcAttr.CreationFlags)
}
