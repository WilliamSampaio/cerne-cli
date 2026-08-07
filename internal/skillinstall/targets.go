package skillinstall

import (
	"errors"
	"path/filepath"
	"strings"
)

const (
	SkillName   = "cerne-context"
	PackageName = "cerne-skills"
)

var ErrInvalidAgent = errors.New("invalid agent")

func SupportedAgent(agent string) bool {
	return agent == "codex" || agent == "claude"
}

func TargetPath(home, agent string) (string, error) {
	if !SupportedAgent(agent) {
		return "", ErrInvalidAgent
	}
	parts := map[string][]string{
		"codex":  {".codex", "skills", SkillName},
		"claude": {".claude", "skills", SkillName},
	}[agent]
	target := filepath.Join(append([]string{home}, parts...)...)
	if inside, err := pathInside(home, target); err != nil || !inside {
		return "", errors.New("destination outside home")
	}
	return target, nil
}

func pathInside(parent, child string) (bool, error) {
	parent, err := filepath.Abs(parent)
	if err != nil {
		return false, err
	}
	child, err = filepath.Abs(child)
	if err != nil {
		return false, err
	}
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false, err
	}
	return rel == "." || rel != ".." && !filepath.IsAbs(rel) && !strings.HasPrefix(rel, ".."+string(filepath.Separator)), nil
}
