//go:build darwin

package workspace

import "golang.org/x/sys/unix"

func promoteSource(staging, source string) error {
	return unix.RenamexNp(staging, source, unix.RENAME_EXCL)
}
