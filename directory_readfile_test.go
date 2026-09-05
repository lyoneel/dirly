package dirly

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ============================================================================
// ReadFileBytes Tests - Success Cases
// ============================================================================

func TestDirectory_ReadFileBytes_Success(t *testing.T) {
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
			testFile := filepath.Join(tmpDir, tt.filename)
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			dir := NewDirectory(tmpDir)

			result, err := dir.ReadToBytes(tt.filename)
			if err != nil {
				t.Errorf("ReadToBytes() unexpected error: %v", err)
			}

			if string(result) != tt.content {
				t.Errorf("expected content %q, got %q", tt.content, string(result))
			}
		})
	}
}

// ============================================================================
// ReadFileBytes Tests - Error Cases
// ============================================================================

func TestDirectory_ReadFileBytes_Error(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "empty filename",
			filename: "",
		},
		{
			name:     "non-existent file",
			filename: "nonexistent.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			result, err := dir.ReadToBytes(tt.filename)
			if err == nil {
				t.Error("ReadToBytes() should return error")
			}

			if result != nil {
				t.Errorf("expected nil result on error, got %v", result)
			}
		})
	}
}

// ============================================================================
// ReadFileStr Tests - Success Cases
// ============================================================================

func TestDirectory_ReadFileStr_Success(t *testing.T) {
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
		{
			name:     "file with newlines",
			filename: "multiline.txt",
			content:  "line1\nline2\nline3",
		},
		{
			name:     "empty file",
			filename: "empty.txt",
			content:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, tt.filename)
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			dir := NewDirectory(tmpDir)

			result, err := dir.ReadToString(tt.filename)
			if err != nil {
				t.Errorf("ReadToString() unexpected error: %v", err)
			}

			if result != tt.content {
				t.Errorf("expected content %q, got %q", tt.content, result)
			}
		})
	}
}

// ============================================================================
// ReadFileStr Tests - Error Cases
// ============================================================================

func TestDirectory_ReadFileStr_Error(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "empty filename",
			filename: "",
		},
		{
			name:     "non-existent file",
			filename: "nonexistent.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			result, err := dir.ReadToString(tt.filename)
			if err == nil {
				t.Error("ReadToString() should return error")
			}

			if result != "" {
				t.Errorf("expected empty string on error, got %q", result)
			}
		})
	}
}

// ============================================================================
// ReadFileLines Tests - Success Cases
// ============================================================================

func TestDirectory_ReadFileLines_Success(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
		content  string
		expected []string
	}{
		{
			name:     "single line",
			filename: "single.txt",
			content:  "hello world",
			expected: []string{"hello world"},
		},
		{
			name:     "multiple lines with newlines",
			filename: "multiline.txt",
			content:  "line1\nline2\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "empty file",
			filename: "empty.txt",
			content:  "",
			expected: []string{},
		},
		{
			name:     "file ending with newline",
			filename: "trailing.txt",
			content:  "line1\nline2\n",
			expected: []string{"line1", "line2"},
		},
		{
			name:     "single line with trailing newline",
			filename: "single_trail.txt",
			content:  "hello\n",
			expected: []string{"hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, tt.filename)
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			dir := NewDirectory(tmpDir)

			result, err := dir.ReadToLines(tt.filename)
			if err != nil {
				t.Errorf("ReadToLines() unexpected error: %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d lines, got %d", len(tt.expected), len(result))
				return
			}

			for i, line := range result {
				if line != tt.expected[i] {
					t.Errorf("line %d: expected %q, got %q", i, tt.expected[i], line)
				}
			}
		})
	}
}

// ============================================================================
// ReadFileLines Tests - Error Cases
// ============================================================================

func TestDirectory_ReadFileLines_Error(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "empty filename",
			filename: "",
		},
		{
			name:     "non-existent file",
			filename: "nonexistent.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			result, err := dir.ReadToLines(tt.filename)
			if err == nil {
				t.Error("ReadToLines() should return error")
			}

			if result != nil && len(result) > 0 {
				t.Errorf("expected nil or empty slice on error, got %v", result)
			}
		})
	}
}

// ============================================================================
// ReadFileBytes, ReadFileStr, and ReadFileLines Consistency Tests
// ============================================================================

func TestDirectory_ReadFileAllConsistency(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "simple content",
			filename: "test.txt",
			content:  "hello world",
		},
		{
			name:     "multiline content",
			filename: "multiline.txt",
			content:  "line1\nline2\nline3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, tt.filename)
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			dir := NewDirectory(tmpDir)

			bytesResult, bytesErr := dir.ReadToBytes(tt.filename)
			strResult, strErr := dir.ReadToString(tt.filename)
			linesResult, linesErr := dir.ReadToLines(tt.filename)

			if bytesErr != nil {
				t.Errorf("ReadToBytes() unexpected error: %v", bytesErr)
			}
			if strErr != nil {
				t.Errorf("ReadToString() unexpected error: %v", strErr)
			}
			if linesErr != nil {
				t.Errorf("ReadToLines() unexpected error: %v", linesErr)
			}

			// Verify bytes and string are consistent
			if string(bytesResult) != strResult {
				t.Errorf("ReadToBytes and ReadToString returned different results")
			}

			// Verify lines can be reconstructed to original string
			reconstructed := strings.Join(linesResult, "\n")
			if tt.content != reconstructed && !(tt.content == reconstructed+"\n" || tt.content == "") {
				t.Errorf("ReadToLines reconstruction mismatch: expected %q, got %q", tt.content, reconstructed)
			}

			// Verify string can be split to get lines (accounting for trailing newline)
			expectedLines := strings.Split(strResult, "\n")
			if len(expectedLines) > 0 && expectedLines[len(expectedLines)-1] == "" {
				expectedLines = expectedLines[:len(expectedLines)-1]
			}

			if len(linesResult) != len(expectedLines) {
				t.Errorf("ReadToLines count mismatch: expected %d, got %d", len(expectedLines), len(linesResult))
			}
		})
	}
}
