//go:build !windows

package filecheck

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

func probe(path string, read bool) (Outcome, string) {
	mode := uint32(unix.W_OK)
	if read {
		mode = unix.R_OK
	}
	if info, err := os.Lstat(path); err == nil && info.IsDir() {
		mode |= unix.X_OK
	}
	err := unix.Faccessat(unix.AT_FDCWD, path, mode, unix.AT_EACCESS)
	if err == nil {
		return Allowed, ""
	}
	if errors.Is(err, unix.EACCES) || errors.Is(err, unix.EPERM) {
		return Denied, "acesso negado"
	}
	return Unknown, "não foi possível confirmar permissões"
}
