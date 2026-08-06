//go:build linux

package workspace

import "golang.org/x/sys/unix"

func promoteDirectoryNoReplace(staging, target string) error {
	return unix.Renameat2(unix.AT_FDCWD, staging, unix.AT_FDCWD, target, unix.RENAME_NOREPLACE)
}
