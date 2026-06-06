package pkg

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
)

func createTestImage(t *testing.T, path string, width, height int, gradient bool) {
	t.Helper()
	
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if gradient {
				// Create gradient pattern (sharper edges)
				r := uint8(x * 255 / width)
				g := uint8(y * 255 / height)
				img.Set(x, y, color.RGBA{r, g, 128, 255})
			} else {
				// Create uniform gray (less sharp)
				img.Set(x, y, color.RGBA{128, 128, 128, 255})
			}
		}
	}
	
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	
	jpeg.Encode(f, img, nil)
}

func TestCalculateImageQualityScore(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "step3_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create sharp image (gradient has high frequency content)
	sharpPath := filepath.Join(tmpDir, "sharp.jpg")
	createTestImage(t, sharpPath, 100, 100, true)
	
	// Create flat image (uniform gray has low frequency content)
	flatPath := filepath.Join(tmpDir, "flat.jpg")
	createTestImage(t, flatPath, 100, 100, false)
	
	sharpScore := calculateImageQualityScore(sharpPath)
	flatScore := calculateImageQualityScore(flatPath)
	
	if sharpScore <= 0 {
		t.Error("Sharp image score should be positive")
	}
	if flatScore <= 0 {
		t.Error("Flat image score should be positive")
	}
	if sharpScore <= flatScore {
		t.Errorf("Sharp image (%f) should score higher than flat image (%f)", sharpScore, flatScore)
	}
}

func TestCalculateImageQualityScoreInvalidPath(t *testing.T) {
	score := calculateImageQualityScore("/nonexistent/path.jpg")
	if score != 0 {
		t.Errorf("Score for nonexistent file should be 0, got %f", score)
	}
}

func TestCalculateImageQualityScoreInvalidFormat(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "step3_test_invalid")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create a non-image file
	txtPath := filepath.Join(tmpDir, "notanimage.jpg")
	os.WriteFile(txtPath, []byte("not an image"), 0644)
	
	score := calculateImageQualityScore(txtPath)
	if score != 0 {
		t.Errorf("Score for invalid image should be 0, got %f", score)
	}
}

func TestIsDirEmpty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "step3_test_empty")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Empty directory should be empty
	if !isDirEmpty(tmpDir, "JPG") {
		t.Error("Empty directory should be reported as empty")
	}
	
	// Add a JPG file
	os.WriteFile(filepath.Join(tmpDir, "test.JPG"), []byte("jpg"), 0644)
	
	// Now should not be empty
	if isDirEmpty(tmpDir, "JPG") {
		t.Error("Directory with JPG should not be empty")
	}
	
	// But should be empty for RAW type
	if !isDirEmpty(tmpDir, "RAW") {
		t.Error("Directory without RAW files should be empty for RAW type")
	}
}

func TestRemoveEmptyDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "step3_test_remove")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create empty subdirectory
	subDir := filepath.Join(tmpDir, "empty")
	os.MkdirAll(subDir, 0755)
	
	// Should remove empty directory
	if !removeEmptyDir(subDir, "JPG") {
		t.Error("Should have removed empty directory")
	}
	
	// Verify it's gone
	if _, err := os.Stat(subDir); !os.IsNotExist(err) {
		t.Error("Directory should have been removed")
	}
}

func TestRemoveNonEmptyDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "step3_test_nonempty")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create directory with file
	subDir := filepath.Join(tmpDir, "notempty")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "test.JPG"), []byte("jpg"), 0644)
	
	// Should NOT remove non-empty directory
	if removeEmptyDir(subDir, "JPG") {
		t.Error("Should not have removed non-empty directory")
	}
	
	// Verify it still exists
	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		t.Error("Directory should still exist")
	}
}
