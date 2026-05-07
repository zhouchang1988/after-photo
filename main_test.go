package main

import (
	"after_photo/pkg"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFullPipelineSteps1to3 runs steps 1-3 on test data (non-interactive)
func TestFullPipelineSteps1to3(t *testing.T) {
	// Copy test data to a temp dir to avoid modifying originals
	tmpDir, err := os.MkdirTemp("", "after_photo_e2e")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Copy test input files
	inputDir := "test/input"
	entries, err := os.ReadDir(inputDir)
	if err != nil {
		t.Skipf("Test input directory not found: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		src := filepath.Join(inputDir, entry.Name())
		dst := filepath.Join(tmpDir, entry.Name())
		data, err := os.ReadFile(src)
		if err != nil {
			t.Logf("Skipping %s: %v", entry.Name(), err)
			continue
		}
		os.WriteFile(dst, data, 0644)
	}

	// Capture output
	var buf bytes.Buffer
	pkg.SetOutput(&buf)
	pkg.ConfirmCh = nil

	// Step 1: Split by file type
	pkg.Step1(tmpDir)
	output := buf.String()
	if !strings.Contains(output, "步骤1完成") {
		t.Errorf("Step1 failed. Output: %s", output)
	}

	// Verify jpg/raw dirs exist
	if _, err := os.Stat(filepath.Join(tmpDir, "jpg")); os.IsNotExist(err) {
		t.Error("jpg/ directory not created after step1")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "raw")); os.IsNotExist(err) {
		t.Error("raw/ directory not created after step1")
	}

	// Step 2: Group duplicates
	buf.Reset()
	pkg.Step2(tmpDir)
	output = buf.String()
	t.Logf("Step2 output: %s", output)

	// Step 3: Select best
	buf.Reset()
	pkg.Step3(tmpDir)
	output = buf.String()
	t.Logf("Step3 output: %s", output)
}

// TestTUIModelInit tests that the TUI model initializes correctly
func TestTUIModelInit(t *testing.T) {
	m := initialModel()
	if m.state != stateInputDir {
		t.Errorf("Expected initial state to be stateInputDir, got %d", m.state)
	}
	if !m.steps[0] || !m.steps[1] || !m.steps[2] {
		t.Error("Steps 1-3 should be selected by default")
	}
	if m.steps[3] {
		t.Error("Step 4 should NOT be selected by default (it's destructive)")
	}
}

// TestChannelWriter tests the channel-based io.Writer
func TestChannelWriter(t *testing.T) {
	ch := make(chan string, 10)
	cw := newChannelWriter(ch)

	// Write a line with newline
	cw.Write([]byte("hello world\n"))

	select {
	case line := <-ch:
		if line != "hello world" {
			t.Errorf("Expected 'hello world', got '%s'", line)
		}
	default:
		t.Error("Expected to receive line from channel")
	}

	// Write without newline, then flush
	cw.Write([]byte("partial"))
	cw.Flush()

	select {
	case line := <-ch:
		if line != "partial" {
			t.Errorf("Expected 'partial', got '%s'", line)
		}
	default:
		t.Error("Expected to receive flushed content from channel")
	}
}

// TestLogWriter tests the log file writer
func TestLogWriter(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "after_photo_log_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	lw := &logWriter{file: tmpFile}
	lw.Write([]byte("\033[32mtest log line\033[0m\n"))

	// Read back the log file
	tmpFile.Seek(0, 0)
	data, _ := os.ReadFile(tmpFile.Name())
	content := string(data)

	// Should contain timestamp prefix and no ANSI codes
	if strings.Contains(content, "\033[") {
		t.Error("Log file should not contain ANSI codes")
	}
	if !strings.Contains(content, "test log line") {
		t.Error("Log file should contain the log message")
	}
	if !strings.Contains(content, "[") {
		t.Error("Log file should contain timestamp prefix")
	}
}

// TestRemoveANSICodes tests ANSI code removal
func TestRemoveANSICodes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"\033[32mgreen text\033[0m", "green text"},
		{"no ansi", "no ansi"},
		{"\033[1m\033[31mred bold\033[0m", "red bold"},
		{"\033[0m", ""},
	}

	for _, tt := range tests {
		result := removeANSICodes(tt.input)
		if result != tt.expected {
			t.Errorf("removeANSICodes(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
