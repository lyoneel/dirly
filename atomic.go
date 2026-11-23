package dirly

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

// ErrOwnership is wrapped by write errors when the configured default
// ownership (WithDefaultOwnership), or the original owner of a file being
// replaced, cannot be applied. The target file is left untouched in that case.
var ErrOwnership = errors.New("failed to apply file ownership")

const tempFileAttempts = 10

// writeWithOwnership atomically writes data to relative:
//
//  1. the target is resolved through the symlink policy, so writing through an
//     allowed symlink updates the link target and keeps the link;
//  2. data goes to a temporary file in the target's directory, which gets the
//     final mode and ownership and is fsynced;
//  3. the temporary file is renamed over the target and, on POSIX systems, the
//     directory is fsynced (best effort).
//
// A crash leaves either the old or the new content, never a mix; at worst a
// stale ".<name>.tmp-*" file remains. On failure the temporary file is removed
// and the target is not modified.
//
// Mode: a new file gets perm (subject to umask, as before); an existing file
// keeps its current mode, matching the previous in-place behavior.
// Ownership: default ownership, if configured, is applied; otherwise an
// existing file keeps its owner and group. Failures wrap ErrOwnership.
func (d *Directory) writeWithOwnership(relative string, data []byte, perm os.FileMode) error {
	t, err := d.resolveTarget(relative, true, true)
	if err != nil {
		return err
	}
	fsys, name, err := d.openFS(t)
	if err != nil {
		return err
	}
	defer func() { _ = fsys.Close() }()

	parent := filepath.Dir(name)
	if err := fsys.MkdirAll(parent, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	uid, gid := d.defaultUID, d.defaultGID
	mode := perm
	existing, err := fsys.Stat(name)
	switch {
	case err == nil:
		if existing.IsDir() {
			return fmt.Errorf("failed to write file: %q is a directory", relative)
		}
		mode = existing.Mode() & (fs.ModePerm | fs.ModeSetuid | fs.ModeSetgid | fs.ModeSticky)
		if ouid, ogid, ok := fileOwner(existing); ok {
			if uid < 0 {
				uid = ouid
			}
			if gid < 0 {
				gid = ogid
			}
		}
	case !errors.Is(err, fs.ErrNotExist):
		return fmt.Errorf("failed to write file: %w", err)
	}

	tmpName, f, err := createTemp(fsys, parent, filepath.Base(name), perm)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = f.Close()
			_ = fsys.Remove(tmpName)
		}
	}()

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	if existing != nil {
		if err := f.Chmod(mode); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	}
	if err := applyOwnership(f, uid, gid); err != nil {
		return fmt.Errorf("%w on %q: %w", ErrOwnership, relative, err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	if err := fsys.Rename(tmpName, name); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	committed = true
	syncDir(fsys, parent)
	return nil
}

// createTemp exclusively creates a uniquely named temporary file next to the target.
func createTemp(fsys fileSystem, dir, base string, perm os.FileMode) (string, *os.File, error) {
	var lastErr error
	for range tempFileAttempts {
		name := filepath.Join(dir, "."+base+".tmp-"+rand.Text()[:12])
		f, err := fsys.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
		if err == nil {
			return name, f, nil
		}
		if !errors.Is(err, fs.ErrExist) {
			return "", nil, err
		}
		lastErr = err
	}
	return "", nil, lastErr
}

// applyOwnership changes ownership of f when uid or gid is set (>= 0) and
// differs from the current value.
func applyOwnership(f *os.File, uid, gid int) error {
	if uid < 0 && gid < 0 {
		return nil
	}
	if fi, err := f.Stat(); err == nil {
		if cuid, cgid, ok := fileOwner(fi); ok && (uid < 0 || uid == cuid) && (gid < 0 || gid == cgid) {
			return nil
		}
	}
	return f.Chown(uid, gid)
}

// syncDir fsyncs a directory so a rename is durable. Best effort: some file
// systems and Windows do not support it, and the data itself is already synced.
func syncDir(fsys fileSystem, dir string) {
	if runtime.GOOS == "windows" {
		return
	}
	df, err := fsys.Open(dir)
	if err != nil {
		return
	}
	_ = df.Sync()
	_ = df.Close()
}
