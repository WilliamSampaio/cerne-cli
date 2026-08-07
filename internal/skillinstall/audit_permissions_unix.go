//go:build !windows

package skillinstall

import (
	"errors"
	"os"
	"syscall"
)

func secureAuditPath(path string, directory bool) error {
	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || directory != info.IsDir() {
		return errors.New("audit inseguro")
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || int(stat.Uid) != os.Geteuid() {
		return errors.New("audit pertence a outro usuário")
	}
	mode := os.FileMode(0o600)
	if directory {
		mode = 0o700
	}
	return os.Chmod(path, mode)
}
