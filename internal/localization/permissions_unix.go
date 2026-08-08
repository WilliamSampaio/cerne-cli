//go:build !windows

package localization

import (
	"errors"
	"os"
	"syscall"
)

func validateUserPath(path string, _ bool) error {
	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeSymlink != 0 {
		return errors.New("unsafe user configuration path")
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || int(stat.Uid) != os.Geteuid() {
		return errors.New("user configuration belongs to another user")
	}
	return nil
}

func secureUserPath(path string, directory bool) error {
	if err := validateUserPath(path, directory); err != nil {
		return err
	}
	mode := os.FileMode(0o600)
	if directory {
		mode = 0o700
	}
	return os.Chmod(path, mode)
}
