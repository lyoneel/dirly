//go:build unix

package dirly

import (
	"os"
	"syscall"
)

// fileOwner returns the uid and gid of fi when the platform exposes them.
func fileOwner(fi os.FileInfo) (uid, gid int, ok bool) {
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, 0, false
	}
	return int(st.Uid), int(st.Gid), true
}
