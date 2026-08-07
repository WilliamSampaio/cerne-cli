//go:build !windows

package skillinstall

import "os"

func atomicReplaceFile(tempPath, targetPath string) error {
	return os.Rename(tempPath, targetPath)
}
