package skillinstall

import (
	"errors"
	"os"
	"path/filepath"
)

func DefaultPackageDir() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	base := filepath.Dir(executable)
	for _, candidate := range []string{
		filepath.Join(base, PackageName),
		filepath.Join(base, "..", "share", "cerne", PackageName),
	} {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
	}
	return "", errors.New("companion package unavailable")
}
