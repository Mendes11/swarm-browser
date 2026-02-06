package commands

import (
	"testing"

	"github.com/mendes11/swarm-browser/internal/core/models"
)

func TestRunClusterHook_Success(t *testing.T) {
	cluster := models.Cluster{Name: "test"}
	hook := models.Hook{Name: "test-hook", Cmd: "true"}
	cmd := RunClusterHook(cluster, hook)
	msg := cmd()

	if _, ok := msg.(HookSucceeded); !ok {
		t.Errorf("Expected HookSucceeded, got %T", msg)
	}
}

func TestRunClusterHook_Failure(t *testing.T) {
	cluster := models.Cluster{Name: "test"}
	hook := models.Hook{Name: "test-hook", Cmd: "false"}
	cmd := RunClusterHook(cluster, hook)
	msg := cmd()

	failed, ok := msg.(HookFailed)
	if !ok {
		t.Fatalf("Expected HookFailed, got %T", msg)
	}
	if failed.Err == nil {
		t.Error("Expected non-nil error")
	}
}

func TestRunClusterHook_CapturesOutput(t *testing.T) {
	cluster := models.Cluster{Name: "test"}
	hook := models.Hook{Name: "test-hook", Cmd: "echo hello && echo world >&2 && false"}
	cmd := RunClusterHook(cluster, hook)
	msg := cmd()

	failed, ok := msg.(HookFailed)
	if !ok {
		t.Fatalf("Expected HookFailed, got %T", msg)
	}
	if failed.Output == "" {
		t.Error("Expected non-empty output")
	}
}
