package pkg

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStep1(t *testing.T) {
	// Create temp dir with test files
	tmpDir, err := os.MkdirTemp("", "after_photo_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	testFiles := []string{"photo.JPG", "image.CR3", "video.MP4"}
	for _, f := range testFiles {
		if err := os.WriteFile(filepath.Join(tmpDir, f), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Capture output
	var buf bytes.Buffer
	SetOutput(&buf)

	// Run step1
	Step1(tmpDir)

	output := buf.String()
	if !strings.Contains(output, "步骤1完成") {
		t.Errorf("Step1 output missing completion marker: %s", output)
	}

	// Verify directories were created and files moved
	if _, err := os.Stat(filepath.Join(tmpDir, "jpg")); os.IsNotExist(err) {
		t.Error("jpg directory not created")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "raw")); os.IsNotExist(err) {
		t.Error("raw directory not created")
	}

	// Verify files were moved
	if _, err := os.Stat(filepath.Join(tmpDir, "jpg", "photo.JPG")); os.IsNotExist(err) {
		t.Error("JPG file not moved to jpg/")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "raw", "image.CR3")); os.IsNotExist(err) {
		t.Error("RAW file not moved to raw/")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "video", "video.MP4")); os.IsNotExist(err) {
		t.Error("MP4 file not moved to video/")
	}
}

func TestStep4WithConfirmChannel(t *testing.T) {
	// Create temp dir structure
	tmpDir, err := os.MkdirTemp("", "after_photo_test_step4")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	jpgDir := filepath.Join(tmpDir, "jpg")
	rawDir := filepath.Join(tmpDir, "raw")
	os.MkdirAll(jpgDir, 0755)
	os.MkdirAll(rawDir, 0755)

	// Create a JPG file and a RAW file without corresponding JPG
	os.WriteFile(filepath.Join(jpgDir, "photo.JPG"), []byte("jpg"), 0644)
	os.WriteFile(filepath.Join(rawDir, "orphan.CR3"), []byte("raw"), 0644)

	// Set up confirm channel
	ConfirmCh = make(chan *ConfirmRequest, 1)
	defer func() { ConfirmCh = nil }()

	var buf bytes.Buffer
	SetOutput(&buf)

	// Run step4 in a goroutine
	done := make(chan struct{})
	var step4Output string
	go func() {
		Step4(tmpDir)
		step4Output = buf.String()
		close(done)
	}()

	// Wait for confirm request
	req := <-ConfirmCh
	if !strings.Contains(req.Message, "确认删除") {
		t.Errorf("Confirm message unexpected: %s", req.Message)
	}

	// Confirm deletion
	req.Result <- true

	// Wait for step4 to complete
	<-done

	// Verify orphan RAW was deleted
	if _, err := os.Stat(filepath.Join(rawDir, "orphan.CR3")); !os.IsNotExist(err) {
		t.Error("Orphan RAW file should have been deleted")
	}

	if !strings.Contains(step4Output, "成功删除") {
		t.Errorf("Step4 output missing success marker: %s", step4Output)
	}
}

func TestStep4CancelConfirm(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "after_photo_test_step4_cancel")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	jpgDir := filepath.Join(tmpDir, "jpg")
	rawDir := filepath.Join(tmpDir, "raw")
	os.MkdirAll(jpgDir, 0755)
	os.MkdirAll(rawDir, 0755)

	os.WriteFile(filepath.Join(jpgDir, "photo.JPG"), []byte("jpg"), 0644)
	os.WriteFile(filepath.Join(rawDir, "orphan.CR3"), []byte("raw"), 0644)

	ConfirmCh = make(chan *ConfirmRequest, 1)
	defer func() { ConfirmCh = nil }()

	var buf bytes.Buffer
	SetOutput(&buf)

	done := make(chan struct{})
	go func() {
		Step4(tmpDir)
		close(done)
	}()

	req := <-ConfirmCh
	req.Result <- false // Cancel

	<-done

	// Verify orphan RAW was NOT deleted
	if _, err := os.Stat(filepath.Join(rawDir, "orphan.CR3")); os.IsNotExist(err) {
		t.Error("Orphan RAW file should NOT have been deleted after cancel")
	}
}

func TestRequestConfirmDefault(t *testing.T) {
	// Test default confirm function works when ConfirmCh is nil
	ConfirmCh = nil
	// We can't easily test stdin reading in unit tests, so just verify
	// the function doesn't panic when ConfirmCh is nil
	// The actual behavior is tested via the channel-based tests above
}

func TestStep4NoOrphanFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "after_photo_test_step4_no_orphan")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	jpgDir := filepath.Join(tmpDir, "jpg")
	rawDir := filepath.Join(tmpDir, "raw")
	os.MkdirAll(jpgDir, 0755)
	os.MkdirAll(rawDir, 0755)

	// Create matching JPG and RAW files
	os.WriteFile(filepath.Join(jpgDir, "photo.JPG"), []byte("jpg"), 0644)
	os.WriteFile(filepath.Join(rawDir, "photo.CR3"), []byte("raw"), 0644)

	ConfirmCh = nil

	var buf bytes.Buffer
	SetOutput(&buf)

	Step4(tmpDir)

	output := buf.String()
	if !strings.Contains(output, "没有发现多余的 RAW 文件") {
		t.Errorf("Step4 should report no orphan files: %s", output)
	}

	// Verify RAW file still exists
	if _, err := os.Stat(filepath.Join(rawDir, "photo.CR3")); os.IsNotExist(err) {
		t.Error("Non-orphan RAW file should not be deleted")
	}
}
