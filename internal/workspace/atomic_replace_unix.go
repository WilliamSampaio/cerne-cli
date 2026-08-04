//go:build !windows

package workspace

import "os"

func atomicReplaceFile(tempPath, targetPath string) error {
	return os.Rename(tempPath, targetPath)
}
