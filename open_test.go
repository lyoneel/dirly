package dirly

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readAllAndClose(t *testing.T, f *os.File) string {
	t.Helper()
	defer func() { _ = f.Close() }()
	b, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestOpen_ReadsFile(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "a.txt"), "alpha")
	mustWrite(t, filepath.Join(root, "sub", "b.txt"), "beta")

	for _, policy := range []SymlinkPolicy{FollowWithinRoot, FollowAny, NoFollow} {
		d := NewFilteredDirectory(root).WithSymlinkPolicy(policy).Build()
		for name, want := range map[string]string{"a.txt": "alpha", filepath.Join("sub", "b.txt"): "beta"} {
			f, err := d.Open(name)
			if err != nil {
				t.Fatalf("%v: Open(%q): %v", policy, name, err)
			}
			if got := readAllAndClose(t, f); got != want {
				t.Fatalf("%v: Open(%q) read %q, want %q", policy, name, got, want)
			}
		}
	}
}

func TestOpen_EmptyName(t *testing.T) {
	d := NewDirectory(t.TempDir())
	if f, err := d.Open(""); err == nil || f != nil {
		t.Fatalf("Open(\"\") = %v, %v; want nil file and an error", f, err)
	}
}

func TestOpen_MissingFileIsNotExist(t *testing.T) {
	root := t.TempDir()
	for _, policy := range []SymlinkPolicy{FollowWithinRoot, FollowAny, NoFollow} {
		d := NewFilteredDirectory(root).WithSymlinkPolicy(policy).Build()
		_, err := d.Open("missing.txt")
		if !errors.Is(err, os.ErrNotExist) || !os.IsNotExist(err) {
			t.Fatalf("%v: Open(missing) = %v; want errors.Is and os.IsNotExist to report not-exist", policy, err)
		}
	}
}

func TestOpen_MissingRootIsNotExist(t *testing.T) {
	d := NewDirectory(filepath.Join(t.TempDir(), "does-not-exist"))
	_, err := d.Open("a.txt")
	if !errors.Is(err, os.ErrNotExist) || !os.IsNotExist(err) {
		t.Fatalf("Open under a missing root = %v; want not-exist", err)
	}
}

func TestOpen_RejectsTraversalAndAbsolutePaths(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "root")
	mustWrite(t, filepath.Join(root, "inside.txt"), "inside")
	mustWrite(t, filepath.Join(tmp, "outside.txt"), "outside")
	abs := filepath.Join(tmp, "outside.txt")

	for _, policy := range []SymlinkPolicy{FollowWithinRoot, FollowAny, NoFollow} {
		d := NewFilteredDirectory(root).WithSymlinkPolicy(policy).Build()
		for _, name := range []string{"../outside.txt", filepath.Join("sub", "..", "..", "outside.txt"), abs} {
			f, err := d.Open(name)
			if err == nil {
				_ = f.Close()
				t.Fatalf("%v: Open(%q) succeeded; want it refused", policy, name)
			}
			if errors.Is(err, os.ErrNotExist) {
				t.Fatalf("%v: Open(%q) = %v; want a validation error, not not-exist", policy, name, err)
			}
		}
	}
}

func TestOpen_DoesNotApplyBuilderFilters(t *testing.T) {
	// Open matches the Read* helpers, which never applied the extension or
	// pattern filters.
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "notes.txt"), "text")
	d := NewFilteredDirectory(root).WithExtensions("md").Build()

	f, err := d.Open("notes.txt")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	_ = f.Close()
	if _, err := d.ReadToBytes("notes.txt"); err != nil {
		t.Fatalf("ReadToBytes: %v", err)
	}
}

func TestReadHelpers_UnchangedErrorsAndContent(t *testing.T) {
	root := t.TempDir()
	mustWrite(t, filepath.Join(root, "a.txt"), "line1\nline2\n")
	d := NewDirectory(root)

	f, r, err := d.ReadToBuff("a.txt")
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(r)
	_ = f.Close()
	if string(b) != "line1\nline2\n" {
		t.Fatalf("ReadToBuff read %q", b)
	}

	f, sc, err := d.ReadToScanner("a.txt")
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	_ = f.Close()
	if strings.Join(lines, ",") != "line1,line2" {
		t.Fatalf("ReadToScanner lines = %v", lines)
	}

	// Open failures keep the historical "failed to open file" prefix in the
	// streaming helpers and stay detectable with errors.Is.
	for name, call := range map[string]func() error{
		"ReadToBuff":    func() error { _, _, err := d.ReadToBuff("missing.txt"); return err },
		"ReadToScanner": func() error { _, _, err := d.ReadToScanner("missing.txt"); return err },
	} {
		err := call()
		if err == nil || !strings.HasPrefix(err.Error(), "failed to open file: ") || !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("%s(missing) = %v; want a wrapped not-exist error", name, err)
		}
	}
	// ReadToBytes returns the OS error as before.
	if _, err := d.ReadToBytes("missing.txt"); !os.IsNotExist(err) {
		t.Fatalf("ReadToBytes(missing) = %v; want os.IsNotExist", err)
	}
	// Validation errors are returned without the open prefix, as before.
	if _, _, err := d.ReadToBuff("../x"); err == nil || strings.HasPrefix(err.Error(), "failed to open file") {
		t.Fatalf("ReadToBuff(../x) = %v; want the traversal error unprefixed", err)
	}
	if _, _, err := d.ReadToBuff(""); err == nil || err.Error() != "filename cannot be empty" {
		t.Fatalf("ReadToBuff(\"\") = %v", err)
	}
}

func TestOpen_FollowWithinRoot(t *testing.T) {
	root, outside := escapeFixture(t)
	symlinkOrSkip(t, "inside.txt", filepath.Join(root, "alias.txt"))
	symlinkOrSkip(t, filepath.Join(outside, "secret.txt"), filepath.Join(root, "leak.txt"))
	symlinkOrSkip(t, outside, filepath.Join(root, "outdir"))
	d := NewDirectory(root)

	f, err := d.Open("alias.txt")
	if err != nil {
		t.Fatalf("in-root symlink: %v", err)
	}
	if got := readAllAndClose(t, f); got != "inside" {
		t.Fatalf("alias.txt read %q", got)
	}
	for _, name := range []string{"leak.txt", filepath.Join("outdir", "secret.txt")} {
		f, err := d.Open(name)
		if f != nil {
			_ = f.Close()
			t.Fatalf("Open(%q) returned a file", name)
		}
		assertEscape(t, err)
		if os.IsNotExist(err) {
			t.Fatalf("Open(%q): an escape must not look like not-exist", name)
		}
	}
	for name, call := range map[string]func() error{
		"ReadToBuff":    func() error { _, _, err := d.ReadToBuff("leak.txt"); return err },
		"ReadToScanner": func() error { _, _, err := d.ReadToScanner("leak.txt"); return err },
		"ReadToBytes":   func() error { _, err := d.ReadToBytes("leak.txt"); return err },
	} {
		if err := call(); !errors.Is(err, ErrPathEscapesRoot) {
			t.Fatalf("%s(leak) = %v; want ErrPathEscapesRoot", name, err)
		}
	}
}

func TestOpen_NoFollow(t *testing.T) {
	root, outside := escapeFixture(t)
	symlinkOrSkip(t, "inside.txt", filepath.Join(root, "alias.txt"))
	symlinkOrSkip(t, filepath.Join(outside, "secret.txt"), filepath.Join(root, "leak.txt"))
	d := NewFilteredDirectory(root).WithSymlinkPolicy(NoFollow).Build()

	for _, name := range []string{"alias.txt", "leak.txt"} {
		f, err := d.Open(name)
		if f != nil {
			_ = f.Close()
		}
		if !errors.Is(err, ErrSymlinkNotAllowed) {
			t.Fatalf("Open(%q) = %v; want ErrSymlinkNotAllowed", name, err)
		}
		var se *SymlinkError
		if !errors.As(err, &se) {
			t.Fatalf("Open(%q): want *SymlinkError, got %T", name, err)
		}
	}
	f, err := d.Open("inside.txt")
	if err != nil {
		t.Fatalf("plain file under NoFollow: %v", err)
	}
	_ = f.Close()
}

func TestOpen_FollowAnyAllowsEscape(t *testing.T) {
	root, outside := escapeFixture(t)
	symlinkOrSkip(t, filepath.Join(outside, "secret.txt"), filepath.Join(root, "leak.txt"))
	d := NewFilteredDirectory(root).WithSymlinkPolicy(FollowAny).Build()

	f, err := d.Open("leak.txt")
	if err != nil {
		t.Fatalf("FollowAny: %v", err)
	}
	if got := readAllAndClose(t, f); got != "secret" {
		t.Fatalf("leak.txt read %q", got)
	}
}

func TestOpen_SymlinkedBaseDirectory(t *testing.T) {
	tmp := t.TempDir()
	realDir := filepath.Join(tmp, "real")
	mustWrite(t, filepath.Join(realDir, "a.txt"), "via link")
	link := filepath.Join(tmp, "link")
	symlinkOrSkip(t, realDir, link)

	for _, policy := range []SymlinkPolicy{FollowWithinRoot, NoFollow, FollowAny} {
		d := NewFilteredDirectory(link).WithSymlinkPolicy(policy).Build()
		f, err := d.Open("a.txt")
		if err != nil {
			t.Fatalf("%v: Open through symlinked base: %v", policy, err)
		}
		if got := readAllAndClose(t, f); got != "via link" {
			t.Fatalf("%v: read %q", policy, got)
		}
	}
}

func TestOpen_SwapAfterOpenKeepsHandle(t *testing.T) {
	// The returned *os.File stays valid after dirly closes its os.Root.
	root, outside := escapeFixture(t)
	d := NewDirectory(root)
	f, err := d.Open("inside.txt")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "inside.txt")); err != nil {
		t.Fatal(err)
	}
	symlinkOrSkip(t, filepath.Join(outside, "secret.txt"), filepath.Join(root, "inside.txt"))
	if got := readAllAndClose(t, f); got != "inside" {
		t.Fatalf("open handle read %q, want the original content", got)
	}
	_, err = d.Open("inside.txt")
	assertEscape(t, err)
}

func BenchmarkOpen(b *testing.B) {
	root := b.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "a", "b"), 0o755); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "a", "b", "n.md"), []byte("---\ntitle: x\n---\nbody\n"), 0o644); err != nil {
		b.Fatal(err)
	}
	name := filepath.Join("a", "b", "n.md")
	for _, policy := range []SymlinkPolicy{FollowWithinRoot, FollowAny} {
		d := NewFilteredDirectory(root).WithSymlinkPolicy(policy).Build()
		b.Run("Open/"+policy.String(), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				f, err := d.Open(name)
				if err != nil {
					b.Fatal(err)
				}
				_ = f.Close()
			}
		})
		b.Run("ReadToBuff/"+policy.String(), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				f, r, err := d.ReadToBuff(name)
				if err != nil {
					b.Fatal(err)
				}
				_ = r
				_ = f.Close()
			}
		})
	}
}
