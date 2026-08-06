//go:build darwin

package workspace

import "golang.org/x/sys/unix"

func promoteDirectoryNoReplace(staging, target string) error {
	return unix.RenamexNp(staging, target, unix.RENAME_EXCL)
}
