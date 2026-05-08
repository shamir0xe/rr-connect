package services

import (
	"context"
	"os/exec"
)

func mockCmdOK(_ context.Context, _ string, _ ...string) *exec.Cmd {
	return exec.Command("true")
}

func mockCmdErr(_ context.Context, _ string, _ ...string) *exec.Cmd {
	return exec.Command("false")
}
