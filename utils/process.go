package utils

import (
	"fmt"
	"os/exec"
	"strings"
)

const (
	stdout = "stdout"
	stderr = "stderr"
)

func ExecProcessInDir(workdir string, command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	if workdir != "" {
		cmd.Dir = workdir
	}

	// Run the command
	fmt.Printf("%v %v\n", command, strings.Join(args, " "))
	output, err := cmd.Output()

	return strings.TrimSpace(string(output)), err
}

func ExecProcess(command string, args ...string) (string, error) {
	return ExecProcessInDir("", command, args...)
}
