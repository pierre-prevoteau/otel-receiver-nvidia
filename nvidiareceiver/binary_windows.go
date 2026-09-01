// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build windows

package nvidiareceiver // import "github.com/pierre-prevoteau/otel-receiver-nvidia/nvidiareceiver"

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// createNoWindow is CREATE_NO_WINDOW, which runs a console application without
// allocating a console. Without it every scrape flashes a console window when
// the collector itself runs in an interactive session.
const createNoWindow = 0x08000000

// installDirs returns the directories the NVIDIA driver installs nvidia-smi.exe
// into, searched when the executable is not on the PATH. Current drivers place
// it in System32; older ones use the NVSMI folder, which is never on the PATH.
// Collectors running as a Windows service can also inherit a PATH that lacks
// both.
func installDirs() []string {
	var dirs []string
	if root := os.Getenv("SystemRoot"); root != "" {
		dirs = append(dirs, filepath.Join(root, "System32"))
	}
	for _, env := range []string{"ProgramW6432", "ProgramFiles", "ProgramFiles(x86)"} {
		if programFiles := os.Getenv(env); programFiles != "" {
			dirs = append(dirs, filepath.Join(programFiles, "NVIDIA Corporation", "NVSMI"))
		}
	}
	return dirs
}

// resolveBinaryPath turns a bare file name into an absolute path, first through
// the PATH (which appends the PATHEXT extensions) and then through the known
// driver install directories. Anything that already contains a separator is
// used as configured.
func resolveBinaryPath(configured string) (string, error) {
	if filepath.Base(configured) != configured {
		return configured, nil
	}

	if path, err := exec.LookPath(configured); err == nil {
		return path, nil
	}

	name := configured
	if filepath.Ext(name) == "" {
		name += ".exe"
	}

	dirs := installDirs()
	for _, dir := range dirs {
		candidate := filepath.Join(dir, name)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("%q not found on PATH nor in %s: install the NVIDIA driver or point \"binary_path\" at nvidia-smi.exe",
		configured, strings.Join(dirs, ", "))
}

func hideConsoleWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
}
