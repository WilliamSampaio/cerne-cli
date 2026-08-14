package skillinstall

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Options struct {
	HomeDir    string
	PackageDir string
	Skill      string
	Now        func() time.Time
}

type Result struct {
	Agent       string
	Skill       string
	Version     string
	Destination string
	Outcome     string
	AuditPath   string
}

type Failure struct {
	Code       string
	Cause      string
	Correction string
}

func (f Failure) Error() string { return f.Cause }

type marker struct {
	Package string   `json:"package"`
	Version string   `json:"version"`
	Agent   string   `json:"agent"`
	Skill   string   `json:"skill"`
	Files   []string `json:"files"`
}

type auditRecord struct {
	SchemaVersion  int    `json:"schema_version"`
	Operation      string `json:"operation"`
	Agent          string `json:"agent"`
	Skill          string `json:"skill"`
	Package        string `json:"package"`
	PackageVersion string `json:"package_version"`
	Destination    string `json:"destination"`
	Status         string `json:"status"`
	ErrorCode      string `json:"error_code"`
	StartedAt      string `json:"started_at"`
	FinishedAt     string `json:"finished_at,omitempty"`
}

var (
	renameInstallDir = os.Rename
	replaceAuditFile = atomicReplaceFile
)

func Install(agent string, options Options) (Result, error) {
	var result Result
	if !SupportedAgent(agent) {
		return result, ErrInvalidAgent
	}
	skillName := options.Skill
	if skillName == "" {
		skillName = SkillName
	}
	if !SupportsSkill(agent, skillName) {
		return result, failure("skill-missing", "skill não suportada para o agente", "use uma skill oficial suportada")
	}
	home := options.HomeDir
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return result, failure("home-unavailable", "não foi possível localizar o diretório pessoal", "configure um diretório pessoal acessível")
		}
	}
	destination, err := TargetPath(home, agent, skillName)
	if err != nil {
		return result, failure("destination-invalid", "destino do agente inválido", "configure um diretório pessoal acessível")
	}
	result.Agent = agent
	result.Skill = skillName
	result.Destination = destination
	now := time.Now
	if options.Now != nil {
		now = options.Now
	}
	audit, err := startAudit(home, agent, skillName, destination, now)
	if err != nil {
		return result, failure("audit-start-failed", "não foi possível registrar a tentativa de instalação", "verifique a segurança e as permissões de ~/.cerne/audit")
	}
	result.AuditPath = audit.path
	fail := func(err error) (Result, error) {
		var f Failure
		if !errors.As(err, &f) {
			f = failure("install-failed", "não foi possível instalar a skill", "verifique permissões e tente novamente")
		}
		if auditErr := audit.finish("failed", result.Version, f.Code); auditErr != nil {
			return result, failure("audit-finalization-failed", "não foi possível finalizar a auditoria da instalação", "verifique ~/.cerne/audit antes de tentar novamente")
		}
		return result, f
	}

	packageDir := options.PackageDir
	if packageDir == "" {
		packageDir, err = materializeEmbeddedPackage()
		if err != nil {
			return fail(failure("package-unavailable", "pacote oficial cerne-skills incorporado está inacessível", "verifique o diretório temporário e reinstale o Cerne"))
		}
		defer os.RemoveAll(packageDir)
	}
	pkg, err := LoadPackage(packageDir, agent, skillName)
	if err != nil {
		return fail(err)
	}
	result.Version = pkg.Version

	current, exists, err := readMarker(destination, agent)
	if err != nil {
		return fail(err)
	}
	sameContent := false
	if exists && current.Package == PackageName && current.Skill == skillName && current.Agent == agent && current.Version == pkg.Version {
		sameContent = managedContentEqual(destination, pkg, current)
	}
	if sameContent {
		result.Outcome = "already"
		if err := audit.finish("succeeded", pkg.Version, ""); err != nil {
			return result, failure("audit-finalization-failed", "não foi possível finalizar a auditoria da instalação", "verifique ~/.cerne/audit antes de tentar novamente")
		}
		return result, nil
	}
	if exists {
		if err := validateManagedUpgrade(destination, current, pkg.Files); err != nil {
			return fail(err)
		}
	}

	staging, err := stagePackage(pkg, home, destination, agent)
	if err != nil {
		return fail(err)
	}
	defer os.RemoveAll(staging)
	if !exists {
		result.Outcome = "installed"
		if err := promoteAbsent(staging, destination); err != nil {
			return fail(err)
		}
	} else {
		result.Outcome = "upgraded"
		if err := promoteManaged(staging, destination, current); err != nil {
			return fail(err)
		}
	}
	if err := audit.finish("succeeded", pkg.Version, ""); err != nil {
		return result, failure("audit-finalization-failed", "não foi possível finalizar a auditoria da instalação", "verifique ~/.cerne/audit antes de tentar novamente")
	}
	return result, nil
}

func stagePackage(pkg Package, home, destination, agent string) (string, error) {
	parent := filepath.Dir(destination)
	if err := ensureInstallDirectoryUnder(home, parent); err != nil {
		return "", failure("destination-inaccessible", "destino do agente inacessível", "verifique permissões no perfil do agente")
	}
	staging, err := os.MkdirTemp(parent, "."+pkg.ID+"-")
	if err != nil {
		return "", failure("destination-inaccessible", "não foi possível criar staging da instalação", "verifique permissões no perfil do agente")
	}
	for _, rel := range pkg.Files {
		if unsafeRelative(rel) {
			os.RemoveAll(staging)
			return "", failure("unsafe-package", "pacote cerne-skills contém caminho inseguro", "reinstale o Cerne")
		}
		src := filepath.Join(pkg.Root, pkg.Skill, rel)
		dst := filepath.Join(staging, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			os.RemoveAll(staging)
			return "", err
		}
		if err := copyFile(src, dst); err != nil {
			os.RemoveAll(staging)
			return "", err
		}
	}
	m := marker{Package: PackageName, Version: pkg.Version, Agent: agent, Skill: pkg.ID, Files: pkg.Files}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		os.RemoveAll(staging)
		return "", err
	}
	if err := os.WriteFile(filepath.Join(staging, ".cerne-install.json"), append(data, '\n'), 0o600); err != nil {
		os.RemoveAll(staging)
		return "", err
	}
	return staging, nil
}

func ensureInstallDirectoryUnder(home, path string) error {
	home, err := filepath.Abs(home)
	if err != nil {
		return err
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}
	inside, err := pathInside(home, path)
	if err != nil || !inside {
		return errors.New("destination outside home")
	}
	rel, err := filepath.Rel(home, path)
	if err != nil {
		return err
	}
	current := home
	if rel == "." {
		return ensureInstallDirectory(current)
	}
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		current = filepath.Join(current, part)
		if err := ensureInstallDirectory(current); err != nil {
			return err
		}
	}
	return nil
}

func ensureInstallDirectory(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return os.Mkdir(path, 0o755)
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return errors.New("unsafe install directory")
	}
	return nil
}

func readMarker(destination, agent string) (marker, bool, error) {
	info, err := os.Lstat(destination)
	if errors.Is(err, os.ErrNotExist) {
		return marker{}, false, nil
	}
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return marker{}, false, failure("unknown-destination", "destino existente não é gerenciado pelo Cerne", "remova ou mova o conteúdo existente antes de instalar")
	}
	data, err := os.ReadFile(filepath.Join(destination, ".cerne-install.json"))
	if err != nil {
		return marker{}, false, failure("unknown-destination", "destino existente não é gerenciado pelo Cerne", "remova ou mova o conteúdo existente antes de instalar")
	}
	var m marker
	if err := json.Unmarshal(data, &m); err != nil || m.Package != PackageName || !SupportedSkill(m.Skill) || m.Agent == "" {
		return marker{}, false, failure("unknown-destination", "destino existente não é gerenciado pelo Cerne", "remova ou mova o conteúdo existente antes de instalar")
	}
	if m.Agent != agent || !validSemver(m.Version) || len(m.Files) == 0 {
		return marker{}, false, failure("unknown-destination", "destino existente não é gerenciado pelo Cerne", "remova ou mova o conteúdo existente antes de instalar")
	}
	for _, file := range m.Files {
		if unsafeRelative(file) {
			return marker{}, false, failure("unknown-destination", "marcador de instalação contém caminho inseguro", "remova ou mova o conteúdo existente antes de instalar")
		}
	}
	return m, true, nil
}

func promoteAbsent(staging, destination string) error {
	if _, err := os.Lstat(destination); !errors.Is(err, os.ErrNotExist) {
		return failure("unknown-destination", "destino existente não é gerenciado pelo Cerne", "remova ou mova o conteúdo existente antes de instalar")
	}
	if err := renameInstallDir(staging, destination); err != nil {
		return failure("promotion-failed", "não foi possível promover a instalação", "verifique permissões no perfil do agente")
	}
	return nil
}

func promoteManaged(staging, destination string, current marker) error {
	backup := destination + ".cerne-backup-" + fmt.Sprint(os.Getpid()) + "-" + time.Now().Format("20060102150405.000000000")
	if err := renameInstallDir(destination, backup); err != nil {
		return failure("promotion-failed", "não foi possível preparar atualização da skill", "verifique permissões no perfil do agente")
	}
	if err := copyUnknownManagedFiles(backup, staging, current); err != nil {
		_ = renameInstallDir(backup, destination)
		return err
	}
	if err := renameInstallDir(staging, destination); err != nil {
		_ = renameInstallDir(backup, destination)
		return failure("promotion-failed", "não foi possível promover a instalação", "verifique permissões no perfil do agente")
	}
	_ = os.RemoveAll(backup)
	return nil
}

func validateManagedUpgrade(destination string, current marker, nextFiles []string) error {
	managed := fileSet(current.Files)
	next := fileSet(nextFiles)
	return filepath.WalkDir(destination, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return failure("unknown-destination", "destino gerenciado contém link simbólico desconhecido", "mova o conteúdo existente antes de atualizar")
		}
		rel, err := filepath.Rel(destination, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == ".cerne-install.json" || managed[rel] {
			return nil
		}
		if next[rel] {
			return failure("unknown-destination", "conteúdo desconhecido conflita com a atualização da skill", "mova o conteúdo existente antes de atualizar")
		}
		return nil
	})
}

func copyUnknownManagedFiles(source, destination string, current marker) error {
	managed := fileSet(current.Files)
	return filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == ".cerne-install.json" || managed[rel] {
			return nil
		}
		target := filepath.Join(destination, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return copyFile(path, target)
	})
}

func managedContentEqual(destination string, pkg Package, current marker) bool {
	if len(current.Files) != len(pkg.Files) {
		return false
	}
	for _, rel := range pkg.Files {
		if !containsFile(current.Files, rel) {
			return false
		}
		want, err := os.ReadFile(filepath.Join(pkg.Root, pkg.Skill, rel))
		if err != nil {
			return false
		}
		got, err := os.ReadFile(filepath.Join(destination, rel))
		if err != nil || string(got) != string(want) {
			return false
		}
	}
	return true
}

func fileSet(files []string) map[string]bool {
	out := make(map[string]bool, len(files))
	for _, file := range files {
		out[file] = true
	}
	return out
}

func copyFile(source, destination string) error {
	src, err := os.Open(source)
	if err != nil {
		return failure("unsafe-package", "pacote cerne-skills contém arquivo inacessível", "reinstale o Cerne")
	}
	defer src.Close()
	dst, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return err
	}
	return dst.Close()
}

func failure(code, cause, correction string) Failure {
	return Failure{Code: code, Cause: cause, Correction: correction}
}

type audit struct{ path string }

func startAudit(home, agent, skillName, destination string, now func() time.Time) (audit, error) {
	auditDir := filepath.Join(home, ".cerne", "audit")
	if err := ensureAuditDir(filepath.Join(home, ".cerne")); err != nil {
		return audit{}, err
	}
	if err := ensureAuditDir(auditDir); err != nil {
		return audit{}, err
	}
	path := filepath.Join(auditDir, "skill-install-"+randomID()+".json")
	record := auditRecord{
		SchemaVersion: 1, Operation: "skill.install", Agent: agent, Skill: skillName, Package: PackageName,
		Destination: destination, Status: "started", StartedAt: now().UTC().Format(time.RFC3339),
	}
	a := audit{path: path}
	if err := a.write(record); err != nil {
		return audit{}, err
	}
	return a, nil
}

func ensureAuditDir(path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(path, 0o700); err != nil {
			return err
		}
		return secureAuditPath(path, true)
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return errors.New("unsafe audit directory")
	}
	return secureAuditPath(path, true)
}

func (a audit) finish(status, version, code string) error {
	data, err := os.ReadFile(a.path)
	if err != nil {
		return err
	}
	var record auditRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return err
	}
	record.Status = status
	record.PackageVersion = version
	record.ErrorCode = code
	record.FinishedAt = time.Now().UTC().Format(time.RFC3339)
	return a.write(record)
}

func (a audit) write(record auditRecord) error {
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(a.path), ".skill-install-*.tmp")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if _, err := temp.Write(append(data, '\n')); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := secureAuditPath(tempPath, false); err != nil {
		return err
	}
	if err := replaceAuditFile(tempPath, a.path); err != nil {
		return err
	}
	return secureAuditPath(a.path, false)
}

func randomID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return fmt.Sprint(time.Now().UnixNano())
	}
	return hex.EncodeToString(value[:])
}
