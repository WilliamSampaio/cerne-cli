//go:build !windows

package localization

import "os"

func atomicReplaceFile(tempPath, targetPath string) error {
	return os.Rename(tempPath, targetPath)
}
