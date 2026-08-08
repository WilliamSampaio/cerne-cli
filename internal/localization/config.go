package localization

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	ConfigUnsafe  = "config-unsafe"
	ConfigRead    = "config-read-failed"
	ConfigInvalid = "config-invalid"
	ConfigWrite   = "config-write-failed"
)

type Config struct {
	Language Language `json:"language"`
}

type ConfigFailure struct {
	Code string
	Err  error
}

func (failure ConfigFailure) Error() string {
	if failure.Err == nil {
		return failure.Code
	}
	return failure.Code + ": " + failure.Err.Error()
}

func (failure ConfigFailure) Unwrap() error { return failure.Err }

func Path(home string) string {
	return filepath.Join(home, ".cerne", "config.json")
}

func Load(home string) (Config, bool, error) {
	root := filepath.Join(home, ".cerne")
	if err := validateExistingPath(root, true); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, false, nil
		}
		return Config{}, false, ConfigFailure{Code: ConfigUnsafe, Err: err}
	}
	path := Path(home)
	if err := validateExistingPath(path, false); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, false, nil
		}
		return Config{}, false, ConfigFailure{Code: ConfigUnsafe, Err: err}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, false, ConfigFailure{Code: ConfigRead, Err: err}
	}
	config, err := decode(data)
	if err != nil {
		return Config{}, false, ConfigFailure{Code: ConfigInvalid, Err: err}
	}
	return config, true, nil
}

func Set(home string, language Language) error {
	if _, err := Parse(string(language)); err != nil {
		return err
	}
	root := filepath.Join(home, ".cerne")
	if err := ensurePrivateDir(root); err != nil {
		return ConfigFailure{Code: ConfigUnsafe, Err: err}
	}
	path := Path(home)
	if err := validateExistingPath(path, false); err != nil && !errors.Is(err, os.ErrNotExist) {
		return ConfigFailure{Code: ConfigUnsafe, Err: err}
	}
	data, err := json.MarshalIndent(Config{Language: language}, "", "  ")
	if err != nil {
		return ConfigFailure{Code: ConfigWrite, Err: err}
	}
	temp, err := os.CreateTemp(root, ".config-*.tmp")
	if err != nil {
		return ConfigFailure{Code: ConfigWrite, Err: err}
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if _, err := temp.Write(append(data, '\n')); err != nil {
		temp.Close()
		return ConfigFailure{Code: ConfigWrite, Err: err}
	}
	if err := temp.Close(); err != nil {
		return ConfigFailure{Code: ConfigWrite, Err: err}
	}
	if err := secureUserPath(tempPath, false); err != nil {
		return ConfigFailure{Code: ConfigWrite, Err: err}
	}
	if err := atomicReplaceFile(tempPath, path); err != nil {
		return ConfigFailure{Code: ConfigWrite, Err: err}
	}
	if err := secureUserPath(path, false); err != nil {
		return ConfigFailure{Code: ConfigWrite, Err: err}
	}
	return nil
}

func Unset(home string) error {
	root := filepath.Join(home, ".cerne")
	if err := validateExistingPath(root, true); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return ConfigFailure{Code: ConfigUnsafe, Err: err}
	}
	path := Path(home)
	if err := validateExistingPath(path, false); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return ConfigFailure{Code: ConfigUnsafe, Err: err}
	}
	if err := os.Remove(path); err != nil {
		return ConfigFailure{Code: ConfigWrite, Err: err}
	}
	return nil
}

func decode(data []byte) (Config, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var config Config
	if err := decoder.Decode(&config); err != nil {
		return Config{}, err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			err = errors.New("extra JSON value")
		}
		return Config{}, err
	}
	if _, err := Parse(string(config.Language)); err != nil {
		return Config{}, err
	}
	return config, nil
}

func validateExistingPath(path string, directory bool) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || directory != info.IsDir() {
		return fmt.Errorf("unsafe path %q", path)
	}
	return validateUserPath(path, directory)
}

func ensurePrivateDir(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(path, 0o700); err != nil {
			return err
		}
		return secureUserPath(path, true)
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("unsafe path %q", path)
	}
	return secureUserPath(path, true)
}
