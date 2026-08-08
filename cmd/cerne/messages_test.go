package main

import (
	"reflect"
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

func TestLocalizerUsesSelectedCatalog(t *testing.T) {
	if got := (localizer{language: localization.English}).text(messageConfigSet, "en"); got != "Saved language: en\n" {
		t.Fatalf("en = %q", got)
	}
	if got := (localizer{language: localization.PortugueseBrazil}).text(messageConfigSet, "pt-BR"); got != "Idioma salvo: pt-BR\n" {
		t.Fatalf("pt-BR = %q", got)
	}
}
