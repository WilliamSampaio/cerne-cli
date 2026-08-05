//go:build linux

package workspace

import "golang.org/x/sys/unix"

func promoteSource(staging, source string) error {
	return unix.Renameat2(unix.AT_FDCWD, staging, unix.AT_FDCWD, source, unix.RENAME_NOREPLACE)
}
