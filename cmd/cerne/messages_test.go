package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/WilliamSampaio/cerne-cli/internal/localization"
)

func TestMessageCatalogsHaveIdenticalKeys(t *testing.T) {
	english := messageCatalogs[localization.English]
	portuguese := messageCatalogs[localization.PortugueseBrazil]
	if len(english) != len(portuguese) {
		t.Fatalf("catálogos têm tamanhos diferentes: en=%d pt-BR=%d", len(english), len(portuguese))
	}
	for id := range english {
		if _, ok := portuguese[id]; !ok {
			t.Errorf("pt-BR sem %q", id)
		}
	}
	for id := range portuguese {
		if _, ok := english[id]; !ok {
			t.Errorf("en sem %q", id)
		}
	}
}

func TestMessageCatalogFormatsMatch(t *testing.T) {
	for id, english := range englishMessages {
		portuguese := portugueseBrazilMessages[id]
		if !reflect.DeepEqual(formatVerbs(english), formatVerbs(portuguese)) {
			t.Errorf("formatos de %q diferem: en=%v pt-BR=%v", id, formatVerbs(english), formatVerbs(portuguese))
		}
	}
}

func TestFailureCatalogsHaveIdenticalKeys(t *testing.T) {
	if len(englishFailureMessages) != len(portugueseFailureMessages) {
		t.Fatalf("falhas têm tamanhos diferentes: en=%d pt-BR=%d", len(englishFailureMessages), len(portugueseFailureMessages))
	}
	for id := range englishFailureMessages {
		if _, ok := portugueseFailureMessages[id]; !ok {
			t.Errorf("pt-BR sem falha %q", id)
		}
	}
}

func formatVerbs(value string) []byte {
	var verbs []byte
	for index := 0; index < len(value); index++ {
		if value[index] != '%' || index+1 >= len(value) {
			continue
		}
		if value[index+1] == '%' {
			index++
			continue
		}
		for index++; index < len(value); index++ {
			if (value[index] >= 'a' && value[index] <= 'z') || (value[index] >= 'A' && value[index] <= 'Z') {
				verbs = append(verbs, value[index])
				break
			}
		}
	}
	return verbs
}

func TestNoCatalogPresentsAgentAsCanonicalOption(t *testing.T) {
	deprecationKeys := map[messageID]bool{
		"init.agent-deprecated":        true,
		"workflow.agent-deprecated":    true,
		"git.inspect.agent-deprecated": true,
	}
	for _, catalog := range messageCatalogs {
		for id, value := range catalog {
			if deprecationKeys[id] {
				continue
			}
			if strings.Contains(value, "--agent") {
				t.Errorf("mensagem %q apresenta --agent como forma canônica: %q", id, value)
			}
		}
	}
	for _, catalog := range failureCatalogs {
		for id, message := range catalog {
			if strings.Contains(message.Cause, "--agent") || strings.Contains(message.Correction, "--agent") {
				t.Errorf("falha %q apresenta --agent como correção: %#v", id, message)
			}
		}
	}
}

func TestLocalizerUsesSelectedCatalog(t *testing.T) {
	if got := (localizer{language: localization.English}).text(messageConfigSet, "en"); got != "Saved language: en\n" {
		t.Fatalf("en = %q", got)
	}
	if got := (localizer{language: localization.PortugueseBrazil}).text(messageConfigSet, "pt-BR"); got != "Idioma salvo: pt-BR\n" {
		t.Fatalf("pt-BR = %q", got)
	}
}

// TestFailureCodesAreTranslatedInBothLanguages garante que todo código de falha emitido pelo
// domínio tenha texto nos dois idiomas: sem isso, a falha cai no texto genérico e o usuário perde
// a causa e a correção específicas (FR-036).
func TestFailureCodesAreTranslatedInBothLanguages(t *testing.T) {
	domainCodes := map[string][]string{
		"link": {
			"repository-name-invalid", "repository-name-reserved", "repository-name-duplicate",
			"repository-already-registered", "repositories-not-independent", "repositories-overlap",
			"manifest-invalid", "manifest-version-unsupported", "manifest-update-failed",
			"manifest-update-unsafe", "workspace-not-found", "manifest-missing", "knowledge-missing",
			"knowledge-invalid", "git-unavailable", "source-path-missing", "source-path-invalid",
			"source-path-not-found", "source-path-inaccessible", "source-not-directory",
			"source-not-git", "source-bare", "source-no-worktree", "source-not-git-root",
			"source-already-configured",
		},
		"unlink": {
			"repository-not-registered", "repository-not-removable", "manifest-invalid",
			"manifest-version-unsupported", "manifest-update-failed", "manifest-update-unsafe",
			"workspace-not-found", "manifest-missing",
		},
	}
	for language, catalog := range failureCatalogs {
		for domain, codes := range domainCodes {
			for _, code := range codes {
				entry, ok := catalog[domain+"."+code]
				if !ok {
					t.Errorf("%s: %s.%s sem tradução", language, domain, code)
					continue
				}
				if entry.Cause == "" || entry.Correction == "" {
					t.Errorf("%s: %s.%s incompleto: %#v", language, domain, code, entry)
				}
			}
		}
	}
}

func TestRepositoryMessageKeysExistInBothLanguages(t *testing.T) {
	keys := []messageID{
		"repository.current", "repository.previous", "repository.new",
		"unlink.usage", "unlink.removed", "unlink.failure.default",
		"status.state.invalid", "completion.desc.unlink",
		messageUnlinkHelp,
	}
	for language := range messageCatalogs {
		messages := localizer{language: language}
		for _, key := range keys {
			if value, ok := messages.find(key); !ok || value == "" {
				t.Errorf("%s: mensagem %q ausente", language, key)
			}
		}
	}
}
