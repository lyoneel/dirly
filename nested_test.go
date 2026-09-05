package dirly_test

import (
	"gitlab.com/lyoneel/dirly"
	"os"
	"path/filepath"
	"testing"
)

func TestNestedDirectoriesWithExtensionFilter(t *testing.T) {
	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	filesContent := map[string]string{
		"root.md":          `---\ntitle: "Root"\n---\n# Root`,
		"subdir/nested.md": `---\ntitle: "Nested"\ndepends: ["../root.md"]\n---\n# Nested`,
	}

	for name, content := range filesContent {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("WriteFile %s failed: %v", name, err)
		}
	}

	dir := dirly.NewFilteredDirectory(tmpDir).WithExtensions("md").Build()

	relPaths, err := dir.GetAllRel()
	if err != nil {
		t.Fatalf("GetAllRel failed: %v", err)
	}

	t.Logf("Found %d files: %v", len(relPaths), relPaths)

	if len(relPaths) != 2 {
		t.Errorf("Expected 2 files (including nested), got %d", len(relPaths))
	}
}

func TestNestedDirectoriesWithWildcardPattern(t *testing.T) {
	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	filesContent := map[string]string{
		"root.md":          `---\ntitle: "Root"\n---\n# Root`,
		"subdir/nested.md": `---\ntitle: "Nested"\ndepends: ["../root.md"]\n---\n# Nested`,
	}

	for name, content := range filesContent {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("WriteFile %s failed: %v", name, err)
		}
	}

	dir := dirly.NewFilteredDirectory(tmpDir).Include("**/*.md").Build()

	relPaths, err := dir.GetAllRel()
	if err != nil {
		t.Fatalf("GetAllRel failed: %v", err)
	}

	t.Logf("Found %d files: %v", len(relPaths), relPaths)

	if len(relPaths) != 2 {
		t.Errorf("Expected 2 files (including nested), got %d", len(relPaths))
	}
}

func TestNestedDirectoriesWithDoubleStarPattern(t *testing.T) {
	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	filesContent := map[string]string{
		"root.md":          `---\ntitle: "Root"\n---\n# Root`,
		"subdir/nested.md": `---\ntitle: "Nested"\ndepends: ["../root.md"]\n---\n# Nested`,
	}

	for name, content := range filesContent {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("WriteFile %s failed: %v", name, err)
		}
	}

	dir := dirly.NewFilteredDirectory(tmpDir).WithExtensions("md").Build()

	relPaths, err := dir.GetAllRel()
	if err != nil {
		t.Fatalf("GetAllRel failed: %v", err)
	}

	t.Logf("Found %d files with WithExtensions(md):  %v", len(relPaths), relPaths)

	if len(relPaths) != 2 {
		t.Errorf("Expected 2 files (including nested), got %d", len(relPaths))
	}
}

func TestNestedDirectoriesWithRecursiveInclude(t *testing.T) {
	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	filesContent := map[string]string{
		"root.md":          `---\ntitle: "Root"\n---\n# Root`,
		"subdir/nested.md": `---\ntitle: "Nested"\ndepends: ["../root.md"]\n---\n# Nested`,
	}

	for name, content := range filesContent {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("WriteFile %s failed: %v", name, err)
		}
	}

	dir := dirly.NewFilteredDirectory(tmpDir).Include("**/*.md").WithExtensions("md").Build()

	relPaths, err := dir.GetAllRel()
	if err != nil {
		t.Fatalf("GetAllRel failed: %v", err)
	}

	t.Logf("Found %d files with **/*.md + WithExtensions: %v", len(relPaths), relPaths)

	if len(relPaths) != 2 {
		t.Errorf("Expected 2 files (including nested), got %d", len(relPaths))
	}
}

func TestNestedDirectoriesWithSubdirPattern(t *testing.T) {
	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	filesContent := map[string]string{
		"root.md":          `---\ntitle: "Root"\n---\n# Root`,
		"subdir/nested.md": `---\ntitle: "Nested"\ndepends: ["../root.md"]\n---\n# Nested`,
	}

	for name, content := range filesContent {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("WriteFile %s failed: %v", name, err)
		}
	}

	dir := dirly.NewFilteredDirectory(tmpDir).Include("subdir/*.md").Build()

	relPaths, err := dir.GetAllRel()
	if err != nil {
		t.Fatalf("GetAllRel failed: %v", err)
	}

	t.Logf("Found %d files with subdir/*.md: %v", len(relPaths), relPaths)

	if len(relPaths) != 1 {
		t.Errorf("Expected 1 file from subdir, got %d", len(relPaths))
	}
}

func TestNestedDirectoriesWithAllPatterns(t *testing.T) {
	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	filesContent := map[string]string{
		"root.md":          `---\ntitle: "Root"\n---\n# Root`,
		"subdir/nested.md": `---\ntitle: "Nested"\ndepends: ["../root.md"]\n---\n# Nested`,
	}

	for name, content := range filesContent {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("WriteFile %s failed: %v", name, err)
		}
	}

	dir := dirly.NewFilteredDirectory(tmpDir).Include("**/*.md").Build()

	absPaths, err := dir.GetAllAbs()
	if err != nil {
		t.Fatalf("GetAllAbs failed: %v", err)
	}

	t.Logf("Found %d absolute files: %v", len(absPaths), absPaths)

	if len(absPaths) != 2 {
		t.Errorf("Expected 2 files (including nested), got %d", len(absPaths))
	}
}

func TestNestedDirectoriesWithMultipleSubdirs(t *testing.T) {
	tmpDir := t.TempDir()

	subDir1 := filepath.Join(tmpDir, "subdir1")
	nestedSubDir := filepath.Join(subDir1, "nested")
	if err := os.MkdirAll(nestedSubDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	filesContent := map[string]string{
		"root.md":                `---\ntitle: "Root"\n---\n# Root`,
		"subdir1/file1.md":       `---\ntitle: "File1"\n---\n# File1`,
		"subdir2/file2.md":       `---\ntitle: "File2"\n---\n# File2`,
		"subdir1/nested/deep.md": `---\ntitle: "Deep"\n---\n# Deep`,
	}

	for name, content := range filesContent {
		path := filepath.Join(tmpDir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("MkdirAll failed: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("WriteFile %s failed: %v", name, err)
		}
	}

	dir := dirly.NewFilteredDirectory(tmpDir).WithExtensions("md").Build()

	relPaths, err := dir.GetAllRel()
	if err != nil {
		t.Fatalf("GetAllRel failed: %v", err)
	}

	t.Logf("Found %d files across multiple subdirs: %v", len(relPaths), relPaths)

	if len(relPaths) != 4 {
		t.Errorf("Expected 4 files (including nested in multiple dirs), got %d", len(relPaths))
	}
}

func TestNestedDirectoriesWithExcludePattern(t *testing.T) {
	tmpDir := t.TempDir()

	subDir1 := filepath.Join(tmpDir, "subdir1")
	if err := os.MkdirAll(subDir1, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	filesContent := map[string]string{
		"root.md":           `---\ntitle: "Root"\n---\n# Root`,
		"subdir1/file1.md":  `---\ntitle: "File1"\n---\n# File1`,
		"subdir2/ignore.md": `---\ntitle: "Ignore"\n---\n# Ignore`,
	}

	for name, content := range filesContent {
		path := filepath.Join(tmpDir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("MkdirAll failed: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("WriteFile %s failed: %v", name, err)
		}
	}

	dir := dirly.NewFilteredDirectory(tmpDir).WithExtensions("md").Exclude("subdir2/*").Build()

	relPaths, err := dir.GetAllRel()
	if err != nil {
		t.Fatalf("GetAllRel failed: %v", err)
	}

	t.Logf("Found %d files with exclude pattern: %v", len(relPaths), relPaths)

	if len(relPaths) != 2 {
		t.Errorf("Expected 2 files (excluding subdir2), got %d", len(relPaths))
	}
}

func TestNestedDirectoriesWithMixedExtensions(t *testing.T) {
	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	filesContent := map[string]string{
		"root.md":          `---\ntitle: "Root"\n---\n# Root`,
		"subdir/nested.md": `---\ntitle: "Nested"\n---\n# Nested`,
		"readme.txt":       `Plain text file`,
		"subdir/data.json": `{"key": "value"}`,
	}

	for name, content := range filesContent {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("WriteFile %s failed: %v", name, err)
		}
	}

	dir := dirly.NewFilteredDirectory(tmpDir).WithExtensions("md").Build()

	relPaths, err := dir.GetAllRel()
	if err != nil {
		t.Fatalf("GetAllRel failed: %v", err)
	}

	t.Logf("Found %d .md files (excluding other extensions): %v", len(relPaths), relPaths)

	if len(relPaths) != 2 {
		t.Errorf("Expected 2 .md files, got %d", len(relPaths))
	}
}

func TestNestedDirectoriesWithCaseInsensitiveFilter(t *testing.T) {
	tmpDir := t.TempDir()

	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	filesContent := map[string]string{
		"root.md":           `---\ntitle: "Root"\n---\n# Root`,
		"subdir/nested.MD":  `---\ntitle: "Nested Uppercase"\n---\n# Nested`,
		"subdir/another.Md": `---\ntitle: "Mixed Case"\n---\n# Mixed`,
	}

	for name, content := range filesContent {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("WriteFile %s failed: %v", name, err)
		}
	}

	dir := dirly.NewFilteredDirectory(tmpDir).WithExtensions("md").Build()

	relPaths, err := dir.GetAllRel()
	if err != nil {
		t.Fatalf("GetAllRel failed: %v", err)
	}

	t.Logf("Found %d files with case-insensitive filter: %v", len(relPaths), relPaths)

	if len(relPaths) != 3 {
		t.Errorf("Expected 3 .md files (case-insensitive), got %d", len(relPaths))
	}
}
