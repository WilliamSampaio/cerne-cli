package skillinstall

import (
	"errors"
	"path/filepath"
	"strings"
)

const (
	SkillName             = "cerne-context"
	ProductDiscoverySkill = "cerne-product-discovery"
	GitWorkflowSkill      = "cerne-git-workflow"
	PackageName           = "cerne-skills"
)

var ErrInvalidAgent = errors.New("invalid agent")

func SupportedRuntime(agent string) bool {
	return agent == "codex" || agent == "claude" || agent == "gemini"
}

func SupportedSkill(skill string) bool {
	return skill == SkillName || skill == ProductDiscoverySkill || skill == GitWorkflowSkill
}

func SupportedSkills(agent string) []string {
	switch agent {
	case "codex", "claude":
		return []string{SkillName, ProductDiscoverySkill, GitWorkflowSkill}
	case "gemini":
		return []string{GitWorkflowSkill}
	default:
		return nil
	}
}

func SupportsSkill(runtime, skill string) bool {
	for _, supported := range SupportedSkills(runtime) {
		if skill == supported {
			return true
		}
	}
	return false
}

func TargetPath(home, runtime string, skill ...string) (string, error) {
	skillName := SkillName
	if len(skill) == 1 {
		skillName = skill[0]
	}
	if len(skill) > 1 || !SupportedSkill(skillName) {
		return "", errors.New("invalid skill")
	}
	if !SupportedRuntime(runtime) {
		return "", ErrInvalidAgent
	}
	if !SupportsSkill(runtime, skillName) {
		return "", ErrInvalidAgent
	}
	parts := map[string][]string{
		"codex":  {".codex", "skills", skillName},
		"claude": {".claude", "skills", skillName},
		"gemini": {".gemini", "skills", skillName},
	}[runtime]
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
