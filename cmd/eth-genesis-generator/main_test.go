package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunInit(t *testing.T) {
	outDir := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := run([]string{"init", "--out", outDir}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit code %d, stderr: %s", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(outDir, "genesis.yaml")); err != nil {
		t.Fatal(err)
	}
}
