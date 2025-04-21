package vm

import (
	"testing"
)

func TestGetFreePort(t *testing.T) {
	port, err := getFreePort()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if port <= 0 {
		t.Fatalf("expected a valid port number, got %d", port)
	}
}

func TestFilePath(t *testing.T) {
	job := Job{
		Query: JobQuery{
			BasePath: "/tmp",
		},
	}
	path := job.filePath(Spec)
	if path == "" {
		t.Fatalf("expected a valid file path, got empty string")
	}
	if path != "/tmp/spec.json" {
		t.Fatalf("expected file path to be /tmp/spec.json, got %s", path)
	}
}

