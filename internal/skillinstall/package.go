package skillinstall

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const manifestFile = "cerne-skills.json"

var semverPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$`)

type Package struct {
	Root    string
	Version string
	Skill   string
	ID      string
	Files   []string
}

type manifest struct {
	SchemaVersion int             `json:"schemaVersion"`
	Name          string          `json:"name"`
	Version       string          `json:"version"`
	Skills        []manifestSkill `json:"skills"`
}

type manifestSkill struct {
	ID       string                     `json:"id"`
	Source   string                     `json:"source"`
	Entry    string                     `json:"entrypoint"`
	Adapters map[string]json.RawMessage `json:"adapters"`
	Requires struct {
		ContextSchema string `json:"contextSchema"`
	} `json:"requires"`
}

func LoadPackage(root, agent string, skillName ...string) (Package, error) {
	requested := SkillName
	if len(skillName) == 1 {
		requested = skillName[0]
	}
	if len(skillName) > 1 || !SupportedSkill(requested) {
		return Package{}, failure("skill-missing", "skill ausente no pacote cerne-skills", "use uma skill oficial suportada")
	}
	data, err := os.ReadFile(filepath.Join(root, manifestFile))
	if err != nil {
		return Package{}, failure("package-unavailable", "pacote oficial cerne-skills incorporado está inacessível", "verifique o diretório temporário e reinstale o Cerne")
	}
	var doc manifest
	if err := json.Unmarshal(data, &doc); err != nil {
		return Package{}, failure("manifest-invalid", "manifesto do pacote cerne-skills inválido", "reinstale o Cerne")
	}
	if doc.SchemaVersion != 1 || doc.Name != PackageName || !validSemver(doc.Version) {
		return Package{}, failure("manifest-incompatible", "manifesto do pacote cerne-skills incompatível", "atualize ou reinstale o Cerne")
	}
	for _, skill := range doc.Skills {
		if skill.ID != requested {
			continue
		}
		if skill.Source == "" || skill.Entry != "SKILL.md" || skill.Requires.ContextSchema != "cerne.context.v1" {
			return Package{}, failure("manifest-incompatible", "skill incompatível", "atualize ou reinstale o Cerne")
		}
		if _, ok := skill.Adapters[agent]; !ok {
			return Package{}, failure("adapter-missing", "adaptador do agente ausente no pacote cerne-skills", "atualize ou reinstale o Cerne")
		}
		files, err := safeWalk(root, skill.Source)
		if err != nil {
			return Package{}, err
		}
		if !containsFile(files, "SKILL.md") {
			return Package{}, failure("manifest-incompatible", "entrypoint da skill ausente", "atualize ou reinstale o Cerne")
		}
		return Package{Root: root, Version: doc.Version, Skill: filepath.Clean(skill.Source), ID: requested, Files: files}, nil
	}
	return Package{}, failure("skill-missing", "skill ausente no pacote cerne-skills", "atualize ou reinstale o Cerne")
}

func containsFile(files []string, want string) bool {
	for _, file := range files {
		if filepath.ToSlash(file) == want {
			return true
		}
	}
	return false
}

func safeWalk(root, relative string) ([]string, error) {
	if unsafeRelative(relative) {
		return nil, failure("unsafe-package", "pacote cerne-skills contém caminho inseguro", "reinstale o Cerne")
	}
	sourceRoot := filepath.Join(root, filepath.Clean(relative))
	var files []string
	err := filepath.WalkDir(sourceRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return errors.New("symlink")
		}
		rel, err := filepath.Rel(sourceRoot, path)
		if err != nil {
			return err
		}
		if rel == "." || entry.IsDir() {
			return nil
		}
		if !info.Mode().IsRegular() {
			return errors.New("non-regular")
		}
		files = append(files, rel)
		return nil
	})
	if err != nil {
		return nil, failure("unsafe-package", "pacote cerne-skills contém entradas inseguras", "reinstale o Cerne")
	}
	return files, nil
}

func unsafeRelative(path string) bool {
	clean := filepath.Clean(path)
	return path == "" || filepath.IsAbs(path) || clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator))
}

func validSemver(version string) bool {
	return semverPattern.MatchString(version)
}
