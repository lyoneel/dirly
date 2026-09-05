package dirly

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// ReadToScanner Tests - Success Cases
// ============================================================================

func TestDirectory_FileReadScanner_Success(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name      string
		filename  string
		content   string
		wantLines int
	}{
		{
			name:      "simple file",
			filename:  "test.txt",
			content:   "line1\nline2\nline3\n",
			wantLines: 3,
		},
		{
			name:      "file in subdirectory",
			filename:  "subdir/test.txt",
			content:   "hello\nworld\n",
			wantLines: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewDirectory(tmpDir)

			testFile := filepath.Join(tmpDir, tt.filename)
			if err := os.MkdirAll(filepath.Dir(testFile), 0755); err != nil {
				t.Fatalf("failed to create directory: %v", err)
			}

			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			file, scanner, err := dir.ReadToScanner(tt.filename)
			if err != nil {
				t.Errorf("ReadToScanner() unexpected error: %v", err)
			}
			defer file.Close()

			var lines []string
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}

			if err := scanner.Err(); err != nil {
				t.Errorf("scanner error: %v", err)
			}

			if len(lines) != tt.wantLines {
				t.Errorf("expected %d lines, got %d", tt.wantLines, len(lines))
			}
		})
	}
}

// ============================================================================
// ReadToScanner Tests - Error Cases
// ============================================================================

func TestDirectory_FileReadScanner_Error(t *testing.T) {
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

			file, scanner, err := dir.ReadToScanner(tt.filename)
			if err == nil {
				t.Error("ReadToScanner() should return error")
			}
			if file != nil || scanner != nil {
				t.Error("ReadToScanner() should return nil on error")
			}
		})
	}
}

// ============================================================================
// ReadToScanner Integration Tests
// ============================================================================

func TestDirectory_FileReadScanner_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()

	dir := NewDirectory(tmpDir)

	testFile := filepath.Join(tmpDir, "empty.txt")
	if err := os.WriteFile(testFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	file, scanner, err := dir.ReadToScanner("empty.txt")
	if err != nil {
		t.Errorf("ReadToScanner() unexpected error: %v", err)
	}
	defer file.Close()

	count := 0
	for scanner.Scan() {
		count++
	}

	if count != 0 {
		t.Errorf("expected 0 lines for empty file, got %d", count)
	}
}

func TestDirectory_FileReadScanner_SingleLine(t *testing.T) {
	tmpDir := t.TempDir()

	dir := NewDirectory(tmpDir)

	testFile := filepath.Join(tmpDir, "single.txt")
	if err := os.WriteFile(testFile, []byte("only line"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	file, scanner, err := dir.ReadToScanner("single.txt")
	if err != nil {
		t.Errorf("ReadToScanner() unexpected error: %v", err)
	}
	defer file.Close()

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(lines) != 1 {
		t.Errorf("expected 1 line, got %d", len(lines))
	} else if lines[0] != "only line" {
		t.Errorf("expected 'only line', got '%s'", lines[0])
	}
}

func TestDirectory_FileReadScanner_WithSpecialChars(t *testing.T) {
	tmpDir := t.TempDir()

	dir := NewDirectory(tmpDir)

	testFile := filepath.Join(tmpDir, "special.txt")
	content := "line with spaces\nline\twith\ttabs\nunicode: 你好世界\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	file, scanner, err := dir.ReadToScanner("special.txt")
	if err != nil {
		t.Errorf("ReadToScanner() unexpected error: %v", err)
	}
	defer file.Close()

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
}
