//go:build !unix

package dirly

import "os"

// fileOwner is not available on this platform.
func fileOwner(os.FileInfo) (uid, gid int, ok bool) { return 0, 0, false }
