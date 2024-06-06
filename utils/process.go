package utils

import (
	"log/slog"
	"os/exec"
	"strings"
)

const (
	stdout = "stdout"
	stderr = "stderr"
)

func ExecProcessInDir(logger *slog.Logger, workdir string, command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	if workdir != "" {
		cmd.Dir = workdir
	}

	// Run the command
	logger.Debug("Running command", "cmd", command, "args", args)
	output, err := cmd.Output()

	return strings.TrimSpace(string(output)), err
}

func ExecProcess(logger *slog.Logger, command string, args ...string) (string, error) {
	return ExecProcessInDir(logger, "", command, args...)
}
