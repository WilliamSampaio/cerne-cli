package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrInvalidName       = errors.New("nome de projeto inválido")
	ErrUnsafeDestination = errors.New("destino inseguro")
)

type Result struct {
	Name          string
	KnowledgePath string
	SourcePath    string
}

func Init(parent, name string, initRepository func(string) error) (result Result, err error) {
	if err := ValidateName(name); err != nil {
		return Result{}, err
	}

	root, err := filepath.Abs(filepath.Join(parent, name))
	if err != nil {
		return Result{}, fmt.Errorf("não foi possível resolver o destino: %w", err)
	}

	var created []string
	defer func() {
		if err == nil {
			return
		}
		if rollbackErr := removeCreated(created); rollbackErr != nil {
			err = fmt.Errorf("%w; falha no rollback: %v", err, rollbackErr)
		}
	}()

	if info, err := os.Lstat(root); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return Result{}, fmt.Errorf("%w: %q não é um diretório regular", ErrUnsafeDestination, root)
		}
		empty, err := isEmpty(root)
		if err != nil {
			return Result{}, fmt.Errorf("não foi possível inspecionar o destino %q: %w", root, err)
		}
		if !empty {
			return Result{}, fmt.Errorf("%w: %q não está vazio", ErrUnsafeDestination, root)
		}
	} else if os.IsNotExist(err) {
		if err := os.Mkdir(root, 0o755); err != nil {
			return Result{}, fmt.Errorf("não foi possível criar o destino %q: %w", root, err)
		}
		created = append(created, root)
	} else {
		return Result{}, fmt.Errorf("não foi possível inspecionar o destino %q: %w", root, err)
	}

	knowledge := filepath.Join(root, "knowledge")
	source := filepath.Join(root, "source")
	for _, directory := range append([]string{knowledge, source},
		knowledgeDirectories(knowledge)...) {
		if err := os.Mkdir(directory, 0o755); err != nil {
			return Result{}, fmt.Errorf("não foi possível criar %q: %w", directory, err)
		}
		created = append(created, directory)
	}

	manifestPath := filepath.Join(knowledge, "cerne.json")
	manifest, err := os.OpenFile(manifestPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return Result{}, fmt.Errorf("não foi possível criar o manifesto: %w", err)
	}
	created = append(created, manifestPath)
	encoder := json.NewEncoder(manifest)
	encoder.SetIndent("", "  ")
	writeErr := encoder.Encode(struct {
		Name   string `json:"name"`
		Source string `json:"source"`
	}{name, "../source"})
	closeErr := manifest.Close()
	if writeErr != nil {
		return Result{}, fmt.Errorf("não foi possível gravar o manifesto: %w", writeErr)
	}
	if closeErr != nil {
		return Result{}, fmt.Errorf("não foi possível concluir o manifesto: %w", closeErr)
	}

	for _, repository := range []string{knowledge, source} {
		if err := initRepository(repository); err != nil {
			return Result{}, err
		}
	}

	return Result{Name: name, KnowledgePath: knowledge, SourcePath: source}, nil
}

func ValidateName(name string) error {
	if len(name) == 0 || len(name) > 255 || !asciiAlphaNumeric(name[0]) {
		return fmt.Errorf("%w: use de 1 a 255 caracteres ASCII e comece por letra ou número",
			ErrInvalidName)
	}
	for index := 1; index < len(name); index++ {
		character := name[index]
		if !asciiAlphaNumeric(character) && character != '.' && character != '_' && character != '-' {
			return fmt.Errorf("%w: use somente letras ASCII, números, ponto, hífen ou sublinhado",
				ErrInvalidName)
		}
	}
	if name[len(name)-1] == '.' {
		return fmt.Errorf("%w: o nome não pode terminar em ponto", ErrInvalidName)
	}
	stem := strings.ToUpper(strings.SplitN(name, ".", 2)[0])
	if stem == "CON" || stem == "PRN" || stem == "AUX" || stem == "NUL" ||
		len(stem) == 4 && (strings.HasPrefix(stem, "COM") || strings.HasPrefix(stem, "LPT")) &&
			stem[3] >= '1' && stem[3] <= '9' {
		return fmt.Errorf("%w: %q é reservado nos sistemas suportados", ErrInvalidName, name)
	}
	return nil
}

func asciiAlphaNumeric(character byte) bool {
	return character >= 'a' && character <= 'z' ||
		character >= 'A' && character <= 'Z' ||
		character >= '0' && character <= '9'
}

func removeCreated(paths []string) error {
	var rollbackErr error
	for index := len(paths) - 1; index >= 0; index-- {
		if err := os.RemoveAll(paths[index]); err != nil {
			rollbackErr = errors.Join(rollbackErr, fmt.Errorf("remover %q: %w", paths[index], err))
		}
	}
	return rollbackErr
}

func knowledgeDirectories(root string) []string {
	names := []string{"product", "specs", "decisions", "policies", "runs"}
	directories := make([]string, len(names))
	for index, name := range names {
		directories[index] = filepath.Join(root, name)
	}
	return directories
}

func isEmpty(directory string) (bool, error) {
	handle, err := os.Open(directory)
	if err != nil {
		return false, err
	}
	defer handle.Close()
	_, err = handle.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}
