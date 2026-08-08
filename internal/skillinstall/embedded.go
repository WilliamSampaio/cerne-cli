package skillinstall

import (
	"embed"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed bundle
var embeddedPackage embed.FS

func materializeEmbeddedPackage() (string, error) {
	packageFS, err := fs.Sub(embeddedPackage, "bundle")
	if err != nil {
		return "", err
	}
	root, err := os.MkdirTemp("", ".cerne-skills-")
	if err != nil {
		return "", err
	}
	// ponytail: the tiny bundle is materialized to reuse the hardened filesystem validator;
	// load from fs.FS directly if bundle size ever becomes material.
	if err := fs.WalkDir(packageFS, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
		destination := filepath.Join(root, filepath.FromSlash(path))
		if entry.IsDir() {
			return os.MkdirAll(destination, 0o755)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return errors.New("embedded package contains non-regular entry")
		}
		data, err := fs.ReadFile(packageFS, path)
		if err != nil {
			return err
		}
		return os.WriteFile(destination, data, 0o644)
	}); err != nil {
		os.RemoveAll(root)
		return "", err
	}
	return root, nil
}
