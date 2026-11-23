package dirly

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

func assertNoTempFiles(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp-") {
			t.Fatalf("temporary file left behind: %s", e.Name())
		}
	}
}

func TestAtomicWrite_NoTempFilesAndContent(t *testing.T) {
	root := t.TempDir()
	dir := NewDirectory(root)
	writers := map[string]func(string) error{
		"bytes":  func(n string) error { return dir.WriteFromBytes(n, []byte("data\n"), 0) },
		"string": func(n string) error { return dir.WriteFromString(n, "data\n", 0) },
		"lines":  func(n string) error { return dir.WriteFromLines(n, []string{"data"}, 0) },
	}
	for name, w := range writers {
		for i := range 2 { // create, then replace
			if err := w(name + ".txt"); err != nil {
				t.Fatalf("%s write %d: %v", name, i, err)
			}
		}
		if got := mustRead(t, filepath.Join(root, name+".txt")); got != "data\n" {
			t.Fatalf("%s content = %q", name, got)
		}
	}
	assertNoTempFiles(t, root)
}

func TestAtomicWrite_PreservesExistingMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX modes")
	}
	root := t.TempDir()
	p := filepath.Join(root, "secret.yaml")
	mustWrite(t, p, "old")
	if err := os.Chmod(p, 0600); err != nil {
		t.Fatal(err)
	}
	dir := NewDirectory(root)
	if err := dir.WriteFromString("secret.yaml", "new", 0644); err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0600 {
		t.Fatalf("mode = %o, want 0600 preserved", fi.Mode().Perm())
	}
	// New files get the requested perm (subject to umask).
	if err := dir.WriteFromString("fresh.yaml", "x", 0600); err != nil {
		t.Fatal(err)
	}
	if fi, _ := os.Stat(filepath.Join(root, "fresh.yaml")); fi.Mode().Perm() != 0600 {
		t.Fatalf("new file mode = %o", fi.Mode().Perm())
	}
}

func TestAtomicWrite_FailureLeavesTargetIntact(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "adir", "x"), "x")
	dir := NewDirectory(root)
	if err := dir.WriteFromString("adir", "boom", 0); err == nil {
		t.Fatal("writing over a directory must fail")
	}
	assertNoTempFiles(t, root)
}

func TestOwnershipFailureIsReturned(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chown not supported")
	}
	if os.Geteuid() == 0 {
		t.Skip("root can chown to anything")
	}
	root := t.TempDir()
	p := filepath.Join(root, "conf.yaml")
	mustWrite(t, p, "original")
	dir := NewFilteredDirectory(root).WithDefaultOwnership(0, 0).Build()

	err := dir.WriteFromString("conf.yaml", "changed", 0)
	if !errors.Is(err, ErrOwnership) {
		t.Fatalf("expected ErrOwnership, got %v", err)
	}
	if got := mustRead(t, p); got != "original" {
		t.Fatalf("target modified despite ownership failure: %q", got)
	}
	assertNoTempFiles(t, root)

	if err := dir.WriteFromString("new.yaml", "x", 0); !errors.Is(err, ErrOwnership) {
		t.Fatalf("expected ErrOwnership for new file, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "new.yaml")); !os.IsNotExist(err) {
		t.Fatal("new file created despite ownership failure")
	}
}

func TestBuilderDefaultsNoOwnershipChange(t *testing.T) {
	d := NewFilteredDirectory(t.TempDir()).Build()
	if d.defaultUID != -1 || d.defaultGID != -1 {
		t.Fatalf("default uid/gid = %d/%d, want -1/-1", d.defaultUID, d.defaultGID)
	}
	if err := d.WriteFromString("a.txt", "a", 0); err != nil {
		t.Fatal(err)
	}
}

func TestBuildCopiesFilterConfig(t *testing.T) {
	b := NewFilteredDirectory(t.TempDir()).WithExtensions("yaml")
	d := b.Build()
	b.WithExtensions("json")
	if got := d.FilterConfig().AllowedExtensions; len(got) != 1 || got[0] != "yaml" {
		t.Fatalf("built Directory changed after builder reuse: %v", got)
	}
}

func TestAtomicWrite_ConcurrentWriters(t *testing.T) {
	root := t.TempDir()
	dir := NewDirectory(root)
	var wg sync.WaitGroup
	const n = 16
	errs := make(chan error, n)
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- dir.WriteFromString("shared.txt", fmt.Sprintf("writer-%02d", i), 0)
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	got := mustRead(t, filepath.Join(root, "shared.txt"))
	if !strings.HasPrefix(got, "writer-") || len(got) != len("writer-00") {
		t.Fatalf("torn content: %q", got)
	}
	assertNoTempFiles(t, root)
}
