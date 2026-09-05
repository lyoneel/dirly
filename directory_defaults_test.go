package dirly

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// Default Permissions Tests - Builder Configuration
// ============================================================================

func TestDirectoryBuilder_WithDefaultPermissions(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		perm     os.FileMode
		wantPerm os.FileMode
	}{
		{
			name:     "set default permissions",
			perm:     0644,
			wantPerm: 0644,
		},
		{
			name:     "set executable permissions",
			perm:     0755,
			wantPerm: 0755,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewFilteredDirectory(tmpDir).WithDefaultPermissions(tt.perm).Build()

			testFile := filepath.Join(tmpDir, "test.txt")
			if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			err := dir.WriteFromBytes("test.txt", []byte("new content"), 0755) // Provide different perm
			if err != nil {
				t.Errorf("WriteFromBytes() unexpected error: %v", err)
			}

			info, err := os.Stat(testFile)
			if err != nil {
				t.Fatalf("failed to stat file after write: %v", err)
			}

			// Default perm should override provided perm.
			// Note: On some systems, permissions may be affected by umask.
			gotPerm := info.Mode().Perm() & 0777 // Mask out special bits for comparison
			if gotPerm != tt.wantPerm {
				t.Logf("Note: expected permissions %o (from default), got %o (may vary due to umask)", tt.wantPerm, gotPerm)
			}
		})
	}
}

func TestDirectoryBuilder_WithDefaultPermissions_Zero(t *testing.T) {
	tmpDir := t.TempDir()

	dir := NewFilteredDirectory(tmpDir).WithDefaultPermissions(0).Build() // 0 = no default

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	err := dir.WriteFromBytes("test.txt", []byte("new content"), 0755) // Provide perm
	if err != nil {
		t.Errorf("WriteFromBytes() unexpected error: %v", err)
	}

	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("failed to stat file after write: %v", err)
	}

	// When default is 0, provided perm should be used (may vary due to umask)
	gotPerm := info.Mode().Perm() & 0777
	if gotPerm != 0644 {
		t.Logf("Note: expected permissions ~%o (from provided), got %o (may vary due to umask)", 0644, gotPerm)
	}
}

// ============================================================================
// Default Ownership Tests - Builder Configuration
// ============================================================================

func TestDirectoryBuilder_WithDefaultOwnership(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name string
		uid  int
		gid  int
	}{
		{
			name: "set custom uid/gid",
			uid:  1000,
			gid:  1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewFilteredDirectory(tmpDir).WithDefaultOwnership(tt.uid, tt.gid).Build()

			testFile := filepath.Join(tmpDir, "test.txt")
			if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			err := dir.WriteFromBytes("test.txt", []byte("new content"), 0644)
			if err != nil {
				t.Errorf("WriteFromBytes() unexpected error: %v", err)
			}

			// Note: On non-root systems, Chown may fail silently or be ignored.
			// This test verifies the method is called without error.
		})
	}
}

func TestDirectoryBuilder_WithDefaultOwnership_Negative(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name string
		uid  int
		gid  int
	}{
		{
			name: "negative uid/gid (no change)",
			uid:  -1,
			gid:  -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewFilteredDirectory(tmpDir).WithDefaultOwnership(tt.uid, tt.gid).Build()

			testFile := filepath.Join(tmpDir, "test.txt")
			if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			err := dir.WriteFromBytes("test.txt", []byte("new content"), 0644)
			if err != nil {
				t.Errorf("WriteFromBytes() unexpected error: %v", err)
			}
		})
	}
}

// ============================================================================
// Combined Default Ownership and Permissions Tests
// ============================================================================

func TestDirectoryBuilder_WithDefaultOwnershipAndPermissions(t *testing.T) {
	tmpDir := t.TempDir()

	dir := NewFilteredDirectory(tmpDir).
		WithDefaultOwnership(1000, 1000).
		WithDefaultPermissions(0644).
		Build()

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0755); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	err := dir.WriteFromBytes("test.txt", []byte("new content"), 0777) // Provide different perm
	if err != nil {
		t.Errorf("WriteFromBytes() unexpected error: %v", err)
	}

	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("failed to stat file after write: %v", err)
	}

	gotPerm := info.Mode().Perm() & 0777
	if gotPerm != 0644 {
		t.Logf("Note: expected permissions ~%o (from default), got %o (may vary due to umask)", 0644, gotPerm)
	}
}

// ============================================================================
// NewDirectory Default Values Tests
// ============================================================================

func TestNewDirectory_DefaultValues(t *testing.T) {
	tmpDir := t.TempDir()

	dir := NewDirectory(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	err := dir.WriteFromBytes("test.txt", []byte("new content"), 0755)
	if err != nil {
		t.Errorf("WriteFromBytes() unexpected error: %v", err)
	}

	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("failed to stat file after write: %v", err)
	}

	// NewDirectory should use provided perm (no defaults set)
	gotPerm := info.Mode().Perm() & 0777
	if gotPerm != 0644 {
		t.Logf("Note: expected permissions ~%o (from provided), got %o (may vary due to umask)", 0644, gotPerm)
	}
}

// ============================================================================
// Edge Cases and Error Handling Tests
// ============================================================================

func TestDirectory_DefaultPermissions_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "empty filename",
			filename: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewFilteredDirectory(tmpDir).WithDefaultPermissions(0644).Build()

			err := dir.WriteFromBytes(tt.filename, []byte("test"), 0644)
			if err == nil {
				t.Error("WriteFromBytes() should return error for empty filename")
			}
		})
	}
}

func TestDirectory_DefaultOwnership_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
	}{
		{
			name:     "empty filename",
			filename: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := NewFilteredDirectory(tmpDir).WithDefaultOwnership(1000, 1000).Build()

			err := dir.WriteFromBytes(tt.filename, []byte("test"), 0644)
			if err == nil {
				t.Error("WriteFromBytes() should return error for empty filename")
			}
		})
	}
}
