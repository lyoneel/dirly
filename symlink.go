package dirly

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// SymlinkPolicy controls how a Directory treats symbolic links found inside it.
//
// The directory root itself is always resolved once, when the Directory is
// created (see Directory.ResolvedBasePath), so a root that is itself a symlink
// (for example ~/.config/app -> ~/Sync/app) works under every policy. The
// policy only governs symlinks below the root.
type SymlinkPolicy int

const (
	// FollowWithinRoot follows symlinks inside the directory, but refuses any
	// operation whose fully resolved target lands outside the resolved root.
	// Refusals wrap ErrPathEscapesRoot. This is the default.
	FollowWithinRoot SymlinkPolicy = iota

	// FollowAny follows symlinks wherever they point, including outside the
	// root. This is the behavior of dirly v1.1.0 and earlier. Plain "../" and
	// absolute names are still rejected.
	FollowAny

	// NoFollow refuses any path that goes through a symlink below the root.
	// Refusals wrap ErrSymlinkNotAllowed.
	NoFollow
)

// String returns the policy name.
func (p SymlinkPolicy) String() string {
	switch p {
	case FollowWithinRoot:
		return "FollowWithinRoot"
	case FollowAny:
		return "FollowAny"
	case NoFollow:
		return "NoFollow"
	default:
		return fmt.Sprintf("SymlinkPolicy(%d)", int(p))
	}
}

var (
	// ErrPathEscapesRoot is returned (wrapped in a *SymlinkError) when a path
	// resolves, through symlinks, to a location outside the directory root.
	ErrPathEscapesRoot = errors.New("path escapes directory root")

	// ErrSymlinkNotAllowed is returned (wrapped in a *SymlinkError) when the
	// NoFollow policy is active and a path goes through a symlink.
	ErrSymlinkNotAllowed = errors.New("symlink not allowed by policy")
)

// SymlinkError describes a path refused by the Directory's SymlinkPolicy.
// Use errors.Is with ErrPathEscapesRoot or ErrSymlinkNotAllowed to classify it.
type SymlinkError struct {
	Path     string        // the relative path passed by the caller
	Resolved string        // the resolved location (or the offending link for NoFollow)
	Root     string        // the resolved directory root
	Policy   SymlinkPolicy // the active policy
	Err      error         // ErrPathEscapesRoot or ErrSymlinkNotAllowed
}

func (e *SymlinkError) Error() string {
	if errors.Is(e.Err, ErrSymlinkNotAllowed) {
		return fmt.Sprintf("dirly: %q goes through symlink %q (policy %s): %v", e.Path, e.Resolved, e.Policy, e.Err)
	}
	return fmt.Sprintf("dirly: %q resolves to %q, outside root %q (policy %s): %v", e.Path, e.Resolved, e.Root, e.Policy, e.Err)
}

// Unwrap returns the sentinel error.
func (e *SymlinkError) Unwrap() error { return e.Err }

const maxSymlinkHops = 255

// errLinkFound is used internally by walkSymlinks in no-follow mode.
type errLinkFound struct{ link string }

func (e *errLinkFound) Error() string { return "symlink found: " + e.link }

// resolveRoot returns the absolute, symlink-free form of base.
func resolveRoot(base string) (string, error) {
	abs, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs)
}

// walkSymlinks resolves rel (a clean, relative path) component by component,
// starting at root (an absolute, symlink-free path). Unlike
// filepath.EvalSymlinks it tolerates a missing tail (needed for writes to new
// files) and it resolves dangling symlinks to the path they would create.
//
// If followFinal is false the last component is returned without being
// resolved (lstat semantics, used for removals). If noFollow is true, the first
// symlink encountered is reported with *errLinkFound.
func walkSymlinks(root, rel string, followFinal, noFollow bool) (string, error) {
	if rel == "." || rel == "" {
		return root, nil
	}
	sep := string(filepath.Separator)
	comps := strings.Split(rel, sep)
	cur := root
	hops := 0
	for len(comps) > 0 {
		c := comps[0]
		comps = comps[1:]
		switch c {
		case "", ".":
			continue
		case "..":
			cur = filepath.Dir(cur)
			continue
		}
		next := filepath.Join(cur, c)
		if len(comps) == 0 && !followFinal {
			return next, nil
		}
		fi, err := os.Lstat(next)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ENOTDIR) {
				// Nothing below this point exists, so no further symlinks can
				// be involved. The remainder is joined lexically.
				return filepath.Join(append([]string{next}, comps...)...), nil
			}
			return "", err
		}
		if fi.Mode()&fs.ModeSymlink == 0 {
			cur = next
			continue
		}
		if noFollow {
			return "", &errLinkFound{link: next}
		}
		hops++
		if hops > maxSymlinkHops {
			return "", fmt.Errorf("too many levels of symbolic links resolving %q", next)
		}
		target, err := os.Readlink(next)
		if err != nil {
			return "", err
		}
		target = filepath.FromSlash(target)
		if filepath.IsAbs(target) {
			vol := filepath.VolumeName(target)
			cur = vol + sep
			target = target[len(vol):]
		}
		comps = append(strings.Split(target, sep), comps...)
	}
	return cur, nil
}

// isWithin reports whether p is root or lies below it.
func isWithin(root, p string) bool {
	rel, err := filepath.Rel(root, p)
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

// target is a validated, policy-checked location for one operation.
type target struct {
	lexical string // basePath joined with the cleaned relative name (pre-v1.2 form)
	root    string // resolved root ("" if the root does not exist yet)
	real    string // fully resolved absolute location ("" if root is missing)
	rel     string // real relative to root, for os.Root
}

// resolveTarget validates relative textually (resolvePath), resolves it
// against the real root according to the symlink policy and returns where the
// operation must happen.
//
// followFinal=false keeps the last component unresolved (Remove semantics).
// create=true creates the base directory when it does not exist yet, so the
// root can be resolved (writes historically created missing directories).
func (d *Directory) resolveTarget(relative string, followFinal, create bool) (target, error) {
	lexical, err := d.resolvePath(relative)
	if err != nil {
		return target{}, err
	}
	t := target{lexical: lexical}

	root, err := d.resolvedRoot()
	if err != nil && create && errors.Is(err, fs.ErrNotExist) {
		if mkErr := os.MkdirAll(d.basePath, 0755); mkErr != nil {
			return t, fmt.Errorf("failed to create directory: %w", mkErr)
		}
		root, err = d.resolvedRoot()
	}
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			// Root missing: nothing to follow. Callers surface not-exist.
			return t, nil
		}
		return t, err
	}
	t.root = root

	rel, err := filepath.Rel(d.basePath, lexical)
	if err != nil {
		return t, err
	}
	if rel == "." {
		followFinal = true
	}

	real, err := walkSymlinks(root, rel, followFinal, d.symlinkPolicy == NoFollow)
	if err != nil {
		var lf *errLinkFound
		if errors.As(err, &lf) {
			return t, &SymlinkError{Path: relative, Resolved: lf.link, Root: root, Policy: d.symlinkPolicy, Err: ErrSymlinkNotAllowed}
		}
		return t, err
	}
	t.real = real

	if d.symlinkPolicy != FollowAny {
		if !isWithin(root, real) {
			return t, &SymlinkError{Path: relative, Resolved: real, Root: root, Policy: d.symlinkPolicy, Err: ErrPathEscapesRoot}
		}
		t.rel, _ = filepath.Rel(root, real)
	}
	return t, nil
}

// resolvedRoot returns the root resolved at construction, or resolves it now
// if the directory did not exist when the Directory was created.
func (d *Directory) resolvedRoot() (string, error) {
	if d.realRoot != "" {
		return d.realRoot, nil
	}
	return resolveRoot(d.basePath)
}

// fileSystem is the subset of *os.Root used by dirly. It lets the same code
// run confined to the root (os.Root) or unconfined (FollowAny).
type fileSystem interface {
	Open(name string) (*os.File, error)
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
	Stat(name string) (os.FileInfo, error)
	MkdirAll(name string, perm os.FileMode) error
	Remove(name string) error
	RemoveAll(name string) error
	Rename(oldname, newname string) error
	Chmod(name string, mode os.FileMode) error
	Chown(name string, uid, gid int) error
	Close() error
}

// hostFS is the unconfined file system used by FollowAny. Names are absolute.
type hostFS struct{}

func (hostFS) Open(name string) (*os.File, error) { return os.Open(name) }
func (hostFS) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}
func (hostFS) Stat(name string) (os.FileInfo, error)        { return os.Stat(name) }
func (hostFS) MkdirAll(name string, perm os.FileMode) error { return os.MkdirAll(name, perm) }
func (hostFS) Remove(name string) error                     { return os.Remove(name) }
func (hostFS) RemoveAll(name string) error                  { return os.RemoveAll(name) }
func (hostFS) Rename(oldname, newname string) error         { return os.Rename(oldname, newname) }
func (hostFS) Chmod(name string, mode os.FileMode) error    { return os.Chmod(name, mode) }
func (hostFS) Chown(name string, uid, gid int) error        { return os.Chown(name, uid, gid) }
func (hostFS) Close() error                                 { return nil }

// rootFS wraps *os.Root and maps its escape errors to ErrPathEscapesRoot.
type rootFS struct{ r *os.Root }

func mapRootErr(err error) error {
	if err != nil && strings.Contains(err.Error(), "path escapes from parent") {
		return fmt.Errorf("%w: %w", ErrPathEscapesRoot, err)
	}
	return err
}

func (f rootFS) Open(name string) (*os.File, error) {
	file, err := f.r.Open(name)
	return file, mapRootErr(err)
}
func (f rootFS) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	file, err := f.r.OpenFile(name, flag, perm)
	return file, mapRootErr(err)
}
func (f rootFS) Stat(name string) (os.FileInfo, error) {
	fi, err := f.r.Stat(name)
	return fi, mapRootErr(err)
}
func (f rootFS) MkdirAll(name string, perm os.FileMode) error {
	return mapRootErr(f.r.MkdirAll(name, perm))
}
func (f rootFS) Remove(name string) error    { return mapRootErr(f.r.Remove(name)) }
func (f rootFS) RemoveAll(name string) error { return mapRootErr(f.r.RemoveAll(name)) }
func (f rootFS) Rename(oldname, newname string) error {
	return mapRootErr(f.r.Rename(oldname, newname))
}
func (f rootFS) Chmod(name string, mode os.FileMode) error { return mapRootErr(f.r.Chmod(name, mode)) }
func (f rootFS) Chown(name string, uid, gid int) error {
	return mapRootErr(f.r.Chown(name, uid, gid))
}
func (f rootFS) Close() error { return f.r.Close() }

// openFS returns the file system and the name to use for t.
// FollowWithinRoot and NoFollow go through os.Root opened at the resolved
// root, so the kernel-level (openat-based) containment of os.Root still holds
// if a path component is swapped for a symlink after resolveTarget checked it.
func (d *Directory) openFS(t target) (fileSystem, string, error) {
	if t.root == "" {
		// Root does not exist: use the lexical path so callers get the
		// usual not-exist error from the OS.
		return hostFS{}, t.lexical, nil
	}
	if d.symlinkPolicy == FollowAny {
		return hostFS{}, t.real, nil
	}
	r, err := os.OpenRoot(t.root)
	if err != nil {
		return nil, "", err
	}
	return rootFS{r: r}, t.rel, nil
}
