package dirly

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ============================================================================
// WriteFromBuff Tests - Success Cases
// ============================================================================

func TestDirectory_WriteFromBuff_Success(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "simple file",
			filename: "test.txt",
			content:  "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			reader := bufio.NewReader(strings.NewReader(tt.content))
			err := dir.WriteFromBuff(tt.filename, reader, 0644)
			if err != nil {
				t.Errorf("WriteFromBuff() unexpected error: %v", err)
			}

			testFile := filepath.Join(tmpDir, tt.filename)
			result, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatalf("failed to read written file: %v", err)
			}

			if string(result) != tt.content {
				t.Errorf("expected content %q, got %q", tt.content, string(result))
			}
		})
	}
}

// ============================================================================
// WriteFromBuff Tests - Error Cases
// ============================================================================

func TestDirectory_WriteFromBuff_Error(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "empty filename",
			filename: "",
			content:  "test content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			reader := bufio.NewReader(strings.NewReader(tt.content))
			err := dir.WriteFromBuff(tt.filename, reader, 0644)
			if err == nil {
				t.Error("WriteFromBuff() should return error for empty filename")
			}
		})
	}
}
