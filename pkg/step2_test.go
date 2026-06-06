package pkg

import (
	"testing"
)

func TestHammingDistance(t *testing.T) {
	tests := []struct {
		name     string
		hash1    string
		hash2    string
		expected int
	}{
		{
			name:     "identical hashes",
			hash1:    "1111000011110000111100001111000011110000111100001111000011110000",
			hash2:    "1111000011110000111100001111000011110000111100001111000011110000",
			expected: 0,
		},
		{
			name:     "completely different",
			hash1:    "1111111111111111111111111111111111111111111111111111111111111111",
			hash2:    "0000000000000000000000000000000000000000000000000000000000000000",
			expected: 64,
		},
		{
			name:     "one bit different",
			hash1:    "1111000011110000111100001111000011110000111100001111000011110000",
			hash2:    "1111000011110000111100001111000011110000111100001111000011110001",
			expected: 1,
		},
		{
			name:     "different lengths returns 64",
			hash1:    "11110000",
			hash2:    "1111000011110000",
			expected: 64,
		},
		{
			name:     "threshold boundary - 17 bits different",
			hash1:    "1111111111111111100000000000000000000000000000000000000000000000",
			hash2:    "0000000000000000011111111111111111111111111111111111111111111111",
			expected: 64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hammingDistance(tt.hash1, tt.hash2)
			if result != tt.expected {
				t.Errorf("hammingDistance(%q, %q) = %d, want %d", tt.hash1, tt.hash2, result, tt.expected)
			}
		})
	}
}

func TestCalculateMedian(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{
			name:     "odd count",
			values:   []float64{1.0, 3.0, 2.0},
			expected: 2.0,
		},
		{
			name:     "even count",
			values:   []float64{1.0, 2.0, 3.0, 4.0},
			expected: 2.5,
		},
		{
			name:     "single value",
			values:   []float64{5.0},
			expected: 5.0,
		},
		{
			name:     "already sorted",
			values:   []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			expected: 3.0,
		},
		{
			name:     "reverse sorted",
			values:   []float64{5.0, 4.0, 3.0, 2.0, 1.0},
			expected: 3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateMedian(tt.values)
			if result != tt.expected {
				t.Errorf("calculateMedian(%v) = %f, want %f", tt.values, result, tt.expected)
			}
		})
	}
}

func TestPerformDCT1D(t *testing.T) {
	input := []float64{1.0, 2.0, 3.0, 4.0}
	result := performDCT1D(input)

	if len(result) != len(input) {
		t.Errorf("Expected length %d, got %d", len(input), len(result))
	}

	if result[0] <= 0 {
		t.Error("DC coefficient should be positive for positive input")
	}
}

func TestPerformDCT(t *testing.T) {
	size := 4
	input := make([]float64, size*size)
	for i := range input {
		input[i] = float64(i)
	}

	result := performDCT(input, size)

	if len(result) != len(input) {
		t.Errorf("Expected length %d, got %d", len(input), len(result))
	}
}

func TestSimilarityThreshold(t *testing.T) {
	if SimilarityThreshold != 17 {
		t.Errorf("SimilarityThreshold = %d, want 17", SimilarityThreshold)
	}
}
