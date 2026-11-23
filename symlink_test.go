package dirly

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// symlinkOrSkip creates a symlink or skips the test where symlinks are not
// available (e.g. unprivileged Windows).
func symlinkOrSkip(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks not supported: %v", err)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// escapeFixture builds root/ plus outside/secret.txt and returns both.
func escapeFixture(t *testing.T) (root, outside string) {
	t.Helper()
	tmp := t.TempDir()
	root = filepath.Join(tmp, "root")
	outside = filepath.Join(tmp, "outside")
	mustWrite(t, filepath.Join(root, "inside.txt"), "inside")
	mustWrite(t, filepath.Join(outside, "secret.txt"), "secret")
	return root, outside
}

func assertEscape(t *testing.T, err error) {
	t.Helper()
	if !errors.Is(err, ErrPathEscapesRoot) {
		t.Fatalf("expected ErrPathEscapesRoot, got %v", err)
	}
	var se *SymlinkError
	if !errors.As(err, &se) {
		t.Fatalf("expected *SymlinkError, got %T", err)
	}
}

func TestSymlinkPolicy_DefaultIsFollowWithinRoot(t *testing.T) {
	if p := NewDirectory(t.TempDir()).SymlinkPolicy(); p != FollowWithinRoot {
		t.Fatalf("NewDirectory policy = %v", p)
	}
	if p := NewFilteredDirectory(t.TempDir()).Build().SymlinkPolicy(); p != FollowWithinRoot {
		t.Fatalf("builder policy = %v", p)
	}
	for p, want := range map[SymlinkPolicy]string{FollowWithinRoot: "FollowWithinRoot", FollowAny: "FollowAny", NoFollow: "NoFollow", 9: "SymlinkPolicy(9)"} {
		if p.String() != want {
			t.Errorf("String() = %q, want %q", p.String(), want)
		}
	}
}

func TestFollowWithinRoot_ReadEscapeRefused(t *testing.T) {
	root, outside := escapeFixture(t)
	symlinkOrSkip(t, outside, filepath.Join(root, "dirlink"))
	symlinkOrSkip(t, filepath.Join(outside, "secret.txt"), filepath.Join(root, "filelink"))
	symlinkOrSkip(t, "../outside/secret.txt", filepath.Join(root, "rellink"))
	dir := NewDirectory(root)

	for _, name := range []string{"dirlink/secret.txt", "filelink", "rellink"} {
		t.Run(name, func(t *testing.T) {
			_, err := dir.ReadToString(name)
			assertEscape(t, err)
			_, _, err = dir.ReadToBuff(name)
			assertEscape(t, err)
			_, _, err = dir.ReadToScanner(name)
			assertEscape(t, err)
			_, err = dir.AbsPath(name)
			assertEscape(t, err)
			assertEscape(t, dir.Chmod(name, 0600))
			assertEscape(t, dir.Chown(name, -1, -1))
			if dir.Exists(name) {
				t.Error("Exists() must be false for escaping path")
			}
		})
	}
	if _, err := dir.SubDirectory("dirlink"); !errors.Is(err, ErrPathEscapesRoot) {
		t.Fatalf("SubDirectory: expected ErrPathEscapesRoot, got %v", err)
	}
}

func TestFollowWithinRoot_WriteEscapeRefused(t *testing.T) {
	root, outside := escapeFixture(t)
	symlinkOrSkip(t, filepath.Join(outside, "secret.txt"), filepath.Join(root, "filelink"))
	symlinkOrSkip(t, outside, filepath.Join(root, "dirlink"))
	symlinkOrSkip(t, filepath.Join(outside, "created.txt"), filepath.Join(root, "dangling"))
	dir := NewDirectory(root)

	assertEscape(t, dir.WriteFromString("filelink", "pwned", 0))
	assertEscape(t, dir.WriteFromString("dirlink/secret.txt", "pwned", 0))
	assertEscape(t, dir.WriteFromString("dirlink/new.txt", "pwned", 0))
	assertEscape(t, dir.WriteFromBytes("dangling", []byte("pwned"), 0))
	assertEscape(t, dir.BatchWriteFromString(map[string]string{"filelink": "pwned"}))

	if got := mustRead(t, filepath.Join(outside, "secret.txt")); got != "secret" {
		t.Fatalf("outside file modified: %q", got)
	}
	for _, f := range []string{"new.txt", "created.txt"} {
		if _, err := os.Lstat(filepath.Join(outside, f)); !os.IsNotExist(err) {
			t.Fatalf("%s created outside root", f)
		}
	}
	// The link itself is untouched.
	if fi, err := os.Lstat(filepath.Join(root, "filelink")); err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("link replaced: %v", err)
	}
}

func TestFollowWithinRoot_InRootSymlinksAllowed(t *testing.T) {
	root, _ := escapeFixture(t)
	mustWrite(t, filepath.Join(root, "sub", "real.yaml"), "v1")
	symlinkOrSkip(t, "sub/real.yaml", filepath.Join(root, "rel.yaml"))
	symlinkOrSkip(t, filepath.Join(root, "sub", "real.yaml"), filepath.Join(root, "abs.yaml"))
	symlinkOrSkip(t, "sub", filepath.Join(root, "sublink"))
	symlinkOrSkip(t, "../inside.txt", filepath.Join(root, "sub", "up.txt"))
	dir := NewDirectory(root)

	for _, name := range []string{"rel.yaml", "abs.yaml", "sublink/real.yaml"} {
		got, err := dir.ReadToString(name)
		if err != nil || got != "v1" {
			t.Fatalf("read %s = %q, %v", name, got, err)
		}
	}
	if got, err := dir.ReadToString("sub/up.txt"); err != nil || got != "inside" {
		t.Fatalf("read sub/up.txt = %q, %v", got, err)
	}
	if !dir.Exists("rel.yaml") {
		t.Fatal("Exists(rel.yaml) = false")
	}
	sub, err := dir.SubDirectory("sublink")
	if err != nil {
		t.Fatalf("SubDirectory(sublink): %v", err)
	}
	if got, err := sub.ReadToString("real.yaml"); err != nil || got != "v1" {
		t.Fatalf("sub read = %q, %v", got, err)
	}
	if sub.SymlinkPolicy() != FollowWithinRoot {
		t.Fatal("SubDirectory must inherit policy")
	}
}

func TestWriteThroughSymlinkUpdatesTarget(t *testing.T) {
	root, _ := escapeFixture(t)
	real := filepath.Join(root, "sub", "real.yaml")
	mustWrite(t, real, "v1")
	link := filepath.Join(root, "link.yaml")
	symlinkOrSkip(t, "sub/real.yaml", link)
	dir := NewDirectory(root)

	if err := dir.WriteFromString("link.yaml", "v2", 0); err != nil {
		t.Fatalf("write via link: %v", err)
	}
	if got := mustRead(t, real); got != "v2" {
		t.Fatalf("target = %q, want v2", got)
	}
	fi, err := os.Lstat(link)
	if err != nil || fi.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("link was replaced by a regular file")
	}
	assertNoTempFiles(t, filepath.Join(root, "sub"))
	assertNoTempFiles(t, root)

	// A dangling link pointing inside the root creates its target.
	symlinkOrSkip(t, "sub/new.yaml", filepath.Join(root, "dangling.yaml"))
	if err := dir.WriteFromString("dangling.yaml", "new", 0); err != nil {
		t.Fatalf("write via dangling in-root link: %v", err)
	}
	if got := mustRead(t, filepath.Join(root, "sub", "new.yaml")); got != "new" {
		t.Fatalf("dangling target = %q", got)
	}
}

func TestSymlinkedRootAllowed(t *testing.T) {
	tmp := t.TempDir()
	syncDir := filepath.Join(tmp, "Sync", "app")
	mustWrite(t, filepath.Join(syncDir, "config.yaml"), "synced")
	mustWrite(t, filepath.Join(syncDir, "nested", "a.yaml"), "a")
	mustWrite(t, filepath.Join(tmp, "Sync", "other.yaml"), "sibling")
	base := filepath.Join(tmp, "config", "app")
	if err := os.MkdirAll(filepath.Dir(base), 0755); err != nil {
		t.Fatal(err)
	}
	symlinkOrSkip(t, syncDir, base)
	// Absolute link spelled through the symlinked base path: still inside.
	symlinkOrSkip(t, filepath.Join(base, "config.yaml"), filepath.Join(syncDir, "viabase.yaml"))
	// Link to a sibling of the real root: outside.
	symlinkOrSkip(t, "../other.yaml", filepath.Join(syncDir, "sibling.yaml"))

	for _, policy := range []SymlinkPolicy{FollowWithinRoot, NoFollow, FollowAny} {
		t.Run(policy.String(), func(t *testing.T) {
			dir := NewFilteredDirectory(base).WithSymlinkPolicy(policy).Build()
			if dir.BasePath() != base {
				t.Fatalf("BasePath changed: %q", dir.BasePath())
			}
			resolved, err := dir.ResolvedBasePath()
			want, _ := filepath.EvalSymlinks(syncDir)
			if err != nil || resolved != want {
				t.Fatalf("ResolvedBasePath = %q, %v; want %q", resolved, err, want)
			}
			if got, err := dir.ReadToString("config.yaml"); err != nil || got != "synced" {
				t.Fatalf("read = %q, %v", got, err)
			}
			if err := dir.WriteFromString("nested/b.yaml", "b", 0); err != nil {
				t.Fatalf("write: %v", err)
			}
			if got := mustRead(t, filepath.Join(syncDir, "nested", "b.yaml")); got != "b" {
				t.Fatalf("write landed elsewhere: %q", got)
			}
			files, err := dir.GetAllAbs()
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, f := range files {
				if f == filepath.Join(base, "nested", "a.yaml") {
					found = true
				}
			}
			if !found {
				t.Fatalf("GetAllAbs on symlinked root missing nested/a.yaml in base-path form: %v", files)
			}
			_, viaErr := dir.ReadToString("viabase.yaml")
			_, sibErr := dir.ReadToString("sibling.yaml")
			switch policy {
			case FollowWithinRoot:
				if viaErr != nil {
					t.Fatalf("viabase.yaml should be allowed: %v", viaErr)
				}
				assertEscape(t, sibErr)
			case NoFollow:
				if !errors.Is(viaErr, ErrSymlinkNotAllowed) || !errors.Is(sibErr, ErrSymlinkNotAllowed) {
					t.Fatalf("NoFollow should refuse inner links: %v / %v", viaErr, sibErr)
				}
			case FollowAny:
				if viaErr != nil || sibErr != nil {
					t.Fatalf("FollowAny should allow both: %v / %v", viaErr, sibErr)
				}
			}
		})
	}
}

func TestFollowAny_OldBehavior(t *testing.T) {
	root, outside := escapeFixture(t)
	symlinkOrSkip(t, outside, filepath.Join(root, "dirlink"))
	link := filepath.Join(root, "filelink")
	symlinkOrSkip(t, filepath.Join(outside, "secret.txt"), link)
	dir := NewFilteredDirectory(root).WithSymlinkPolicy(FollowAny).Build()

	if got, err := dir.ReadToString("dirlink/secret.txt"); err != nil || got != "secret" {
		t.Fatalf("read = %q, %v", got, err)
	}
	if err := dir.WriteFromString("filelink", "updated", 0); err != nil {
		t.Fatalf("write: %v", err)
	}
	if got := mustRead(t, filepath.Join(outside, "secret.txt")); got != "updated" {
		t.Fatalf("target = %q", got)
	}
	if fi, _ := os.Lstat(link); fi.Mode()&os.ModeSymlink == 0 {
		t.Fatal("FollowAny write replaced the link")
	}
	// Textual traversal is still rejected.
	if _, err := dir.ReadToString("../outside/secret.txt"); err == nil {
		t.Fatal("../ must still be rejected under FollowAny")
	}
	assertNoTempFiles(t, outside)
}

func TestNoFollow_RefusesSymlinks(t *testing.T) {
	root, _ := escapeFixture(t)
	mustWrite(t, filepath.Join(root, "sub", "real.yaml"), "v1")
	symlinkOrSkip(t, "sub/real.yaml", filepath.Join(root, "link.yaml"))
	symlinkOrSkip(t, "sub", filepath.Join(root, "sublink"))
	dir := NewFilteredDirectory(root).WithSymlinkPolicy(NoFollow).Build()

	for _, name := range []string{"link.yaml", "sublink/real.yaml"} {
		if _, err := dir.ReadToString(name); !errors.Is(err, ErrSymlinkNotAllowed) {
			t.Fatalf("read %s: expected ErrSymlinkNotAllowed, got %v", name, err)
		}
		if err := dir.WriteFromString(name, "x", 0); !errors.Is(err, ErrSymlinkNotAllowed) {
			t.Fatalf("write %s: expected ErrSymlinkNotAllowed, got %v", name, err)
		}
	}
	if got := mustRead(t, filepath.Join(root, "sub", "real.yaml")); got != "v1" {
		t.Fatalf("target modified: %q", got)
	}
	if got, err := dir.ReadToString("sub/real.yaml"); err != nil || got != "v1" {
		t.Fatalf("regular read = %q, %v", got, err)
	}
	if err := dir.WriteFromString("sub/new.yaml", "n", 0); err != nil {
		t.Fatalf("regular write: %v", err)
	}
	// Removing a link is allowed: it never touches the target.
	if err := dir.Remove("link.yaml"); err != nil {
		t.Fatalf("remove link: %v", err)
	}
}

func TestRemoveEscapingLinkRemovesOnlyLink(t *testing.T) {
	root, outside := escapeFixture(t)
	symlinkOrSkip(t, filepath.Join(outside, "secret.txt"), filepath.Join(root, "filelink"))
	symlinkOrSkip(t, outside, filepath.Join(root, "dirlink"))
	dir := NewDirectory(root)

	if err := dir.Remove("filelink"); err != nil {
		t.Fatalf("Remove(filelink): %v", err)
	}
	if err := dir.BatchRemoveAllFromDir("dirlink"); err != nil {
		t.Fatalf("RemoveAll(dirlink): %v", err)
	}
	if got := mustRead(t, filepath.Join(outside, "secret.txt")); got != "secret" {
		t.Fatalf("outside target affected: %q", got)
	}
	// Removing through an escaping directory link is refused.
	symlinkOrSkip(t, outside, filepath.Join(root, "dirlink2"))
	assertEscape(t, dir.Remove("dirlink2/secret.txt"))
	assertEscape(t, dir.BatchRemove([]string{"dirlink2/secret.txt"}))
	if _, err := os.Stat(filepath.Join(outside, "secret.txt")); err != nil {
		t.Fatalf("outside file removed: %v", err)
	}
}

func TestWalkSymlinks_LoopDetected(t *testing.T) {
	root := t.TempDir()
	symlinkOrSkip(t, "b", filepath.Join(root, "a"))
	symlinkOrSkip(t, "a", filepath.Join(root, "b"))
	dir := NewDirectory(root)
	if _, err := dir.ReadToString("a"); err == nil {
		t.Fatal("expected error for symlink loop")
	}
}

func TestIsWithin(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "r")
	cases := map[string]bool{
		root:                           true,
		filepath.Join(root, "a"):       true,
		filepath.Join(root, "..foo"):   true,
		filepath.Join(root, "..", "x"): false,
		filepath.Dir(root):             false,
		root + "x":                     false,
	}
	for p, want := range cases {
		if got := isWithin(root, p); got != want {
			t.Errorf("isWithin(%q) = %v, want %v", p, got, want)
		}
	}
}

func TestNonExistentRoot_WriteCreatesIt(t *testing.T) {
	base := filepath.Join(t.TempDir(), "later", "app")
	dir := NewDirectory(base)
	if _, err := dir.ResolvedBasePath(); err == nil {
		t.Fatal("ResolvedBasePath should fail for missing dir")
	}
	if _, err := dir.ReadToString("x"); err == nil {
		t.Fatal("read from missing dir should fail")
	}
	if err := dir.WriteFromString("x/y.txt", "ok", 0); err != nil {
		t.Fatalf("write: %v", err)
	}
	if got := mustRead(t, filepath.Join(base, "x", "y.txt")); got != "ok" {
		t.Fatalf("got %q", got)
	}
}
