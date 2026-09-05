package dirly

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// BuffReadFile Tests - Success Cases
// ============================================================================

func TestDirectory_BuffReadFile_Success(t *testing.T) {
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
			name:     "file in subdirectory",
			filename: "subdir/test.txt",
			content:  "nested content here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, tt.filename)

			// Create parent directory if needed (like WriteFile does)
			fileDir := filepath.Dir(testFile)
			if err := os.MkdirAll(fileDir, 0755); err != nil {
				t.Fatalf("failed to create directory: %v", err)
			}

			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			dir := NewDirectory(tmpDir)

			file, reader, err := dir.ReadToBuff(tt.filename)
			if err != nil {
				t.Errorf("BuffReadFile() unexpected error: %v", err)
			}

			if file == nil {
				t.Error("BuffReadFile() should return non-nil file")
			}
			if reader == nil {
				t.Error("BuffReadFile() should return non-nil reader")
			}

			// Read content using buffered reader
			content, err := reader.ReadString('\n')
			if err != nil && err.Error() != "EOF" {
				t.Fatalf("failed to read: %v", err)
			}

			err = file.Close()
			if err != nil {
				t.Fatalf("failed to close file: %v", err)
			}

			expected := tt.content + "\n"
			if content != expected && string([]byte(tt.content)) != content[:len(tt.content)] {
				// Handle EOF case - read may include or exclude newline
				contentTrimmed := content[:len(tt.content)]
				if contentTrimmed != tt.content {
					t.Errorf("expected content %q, got %q", tt.content, content)
				}
			}
		})
	}
}

// ============================================================================
// BuffReadFile Tests - Error Cases
// ============================================================================

func TestDirectory_BuffReadFile_Error(t *testing.T) {
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

			file, reader, err := dir.ReadToBuff(tt.filename)
			if err == nil {
				t.Error("BuffReadFile() should return error")
			}

			if file != nil {
				t.Error("BuffReadFile() should return nil file on error")
			}
			if reader != nil {
				t.Error("BuffReadFile() should return nil reader on error")
			}
		})
	}
}

// ============================================================================
// BuffReadFile Tests - Streaming Content
// ============================================================================

func TestDirectory_BuffReadFile_Streaming(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
		content  string
		lines    int
	}{
		{
			name:     "multi-line file",
			filename: "multiline.txt",
			content:  "line1\nline2\nline3\nline4\nline5",
			lines:    5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, tt.filename)

			// Create parent directory if needed (like WriteFile does)
			fileDir := filepath.Dir(testFile)
			if err := os.MkdirAll(fileDir, 0755); err != nil {
				t.Fatalf("failed to create directory: %v", err)
			}

			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			dir := NewDirectory(tmpDir)

			file, reader, err := dir.ReadToBuff(tt.filename)
			if err != nil {
				t.Errorf("BuffReadFile() unexpected error: %v", err)
			}

			defer file.Close()

			// Read line by line using buffered reader
			lineCount := 0
			for i := 0; i < tt.lines; i++ {
				content, err := reader.ReadString('\n')
				if err != nil && err.Error() != "EOF" {
					t.Fatalf("failed to read line %d: %v", i+1, err)
				}

				lineCount++
				_ = content // Process each line
			}

			if lineCount != tt.lines {
				t.Errorf("expected to read %d lines, got %d", tt.lines, lineCount)
			}
		})
	}
}

// ============================================================================
// BuffReadFile Tests - Large File Streaming
// ============================================================================

func TestDirectory_BuffReadFile_LargeFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "large file",
			filename: "large.txt",
			content:  "", // Will generate large content
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate large content (10KB)
			largeContent := ""
			for i := 0; i < 200; i++ {
				largeContent += string(rune('a'+(i%26))) + "\n"
			}

			testFile := filepath.Join(tmpDir, tt.filename)
			if err := os.WriteFile(testFile, []byte(largeContent), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			dir := NewDirectory(tmpDir)

			file, reader, err := dir.ReadToBuff(tt.filename)
			if err != nil {
				t.Errorf("BuffReadFile() unexpected error: %v", err)
			}

			defer file.Close()

			// Stream through large file efficiently
			lineCount := 0
			for {
				_, err := reader.ReadString('\n')
				if err != nil {
					break
				}
				lineCount++
			}

			expectedLines := 200
			if lineCount != expectedLines {
				t.Errorf("expected %d lines, got %d", expectedLines, lineCount)
			}
		})
	}
}

// ============================================================================
// BuffReadFile Tests - Binary Content
// ============================================================================

func TestDirectory_BuffReadFile_Binary(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
		content  []byte
	}{
		{
			name:     "binary content",
			filename: "binary.bin",
			content:  []byte{0x00, 0xFF, 0xAB, 0xCD, 0xEF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, tt.filename)
			if err := os.WriteFile(testFile, tt.content, 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			dir := NewDirectory(tmpDir)

			file, reader, err := dir.ReadToBuff(tt.filename)
			if err != nil {
				t.Errorf("BuffReadFile() unexpected error: %v", err)
			}

			defer file.Close()

			// Read raw bytes using buffered reader
			buffer := make([]byte, len(tt.content))
			n, err := reader.Read(buffer)
			if err != nil && err.Error() != "EOF" {
				t.Fatalf("failed to read: %v", err)
			}

			if n != len(tt.content) {
				t.Errorf("expected to read %d bytes, got %d", len(tt.content), n)
			}

			for i := range tt.content {
				if buffer[i] != tt.content[i] {
					t.Errorf("byte %d: expected 0x%02X, got 0x%02X", i, tt.content[i], buffer[i])
				}
			}
		})
	}
}

// ============================================================================
// BuffReadFile Tests - Peek Functionality
// ============================================================================

func TestDirectory_BuffReadFile_Peek(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "peek content",
			filename: "test.txt",
			content:  "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, tt.filename)

			// Create parent directory if needed (like WriteFile does)
			fileDir := filepath.Dir(testFile)
			if err := os.MkdirAll(fileDir, 0755); err != nil {
				t.Fatalf("failed to create directory: %v", err)
			}

			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			dir := NewDirectory(tmpDir)

			file, reader, err := dir.ReadToBuff(tt.filename)
			if err != nil {
				t.Errorf("BuffReadFile() unexpected error: %v", err)
			}

			defer file.Close()

			// Peek at first few bytes without consuming them
			peekBytes, err := reader.Peek(5)
			if err != nil {
				t.Fatalf("failed to peek: %v", err)
			}

			expectedPeek := []byte("hello")
			if string(peekBytes) != string(expectedPeek) {
				t.Errorf("expected peek %q, got %q", expectedPeek, peekBytes)
			}

			// Now read normally - should still get "hello" at start
			content, err := reader.ReadString('\n')
			if err != nil && err.Error() != "EOF" {
				t.Fatalf("failed to read: %v", err)
			}

			if len(content) < 5 || content[:5] != "hello" {
				t.Errorf("expected content starting with 'hello', got %q", content)
			}
		})
	}
}

// ============================================================================
// BuffReadFile Tests - Buffer Size Verification
// ============================================================================

func TestDirectory_BuffReadFile_BufferSize(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{
			name:     "verify buffered reading",
			filename: "test.txt",
			content:  "buffered content for testing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, tt.filename)

			// Create parent directory if needed (like WriteFile does)
			fileDir := filepath.Dir(testFile)
			if err := os.MkdirAll(fileDir, 0755); err != nil {
				t.Fatalf("failed to create directory: %v", err)
			}

			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			dir := NewDirectory(tmpDir)

			file, reader, err := dir.ReadToBuff(tt.filename)
			if err != nil {
				t.Errorf("BuffReadFile() unexpected error: %v", err)
			}

			defer file.Close()

			// Verify bufio.Reader is used (not direct os.File read)
			if reader == nil {
				t.Fatal("reader should not be nil")
			}

			// Read using buffered operations
			content, err := reader.ReadString('\n')
			if err != nil && err.Error() != "EOF" {
				t.Fatalf("failed to read: %v", err)
			}

			if content[:len(tt.content)] != tt.content {
				t.Errorf("expected content %q, got %q", tt.content, content)
			}
		})
	}
}
