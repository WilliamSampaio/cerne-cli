package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
)

// RepositoryEntry é um repositório de código adicional registrado no manifesto, além do source.
type RepositoryEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// reservedRepositoryNames são os identificadores que o workspace já usa para repositórios próprios.
var reservedRepositoryNames = []string{"source", "knowledge"}

func decodeRepositories(raw map[string]json.RawMessage, target *[]RepositoryEntry) error {
	value, ok := raw["repositories"]
	if !ok {
		return nil
	}
	var entries []RepositoryEntry
	if err := json.Unmarshal(value, &entries); err != nil {
		return errors.New("campo repositories inválido")
	}
	seen := make([]RepositoryEntry, 0, len(entries))
	for index, entry := range entries {
		if err := validateRepositoryName(entry.Name, seen); err != nil {
			return fmt.Errorf("repositories[%d]: %w", index, err)
		}
		if entry.Path == "" {
			return fmt.Errorf("repositories[%d]: campo path inválido", index)
		}
		seen = append(seen, entry)
	}
	*target = entries
	return nil
}

func validateRepositoryName(name string, existing []RepositoryEntry) error {
	if err := ValidateName(name); err != nil {
		return err
	}
	for _, reserved := range reservedRepositoryNames {
		if name == reserved {
			return fmt.Errorf("%q é reservado para um repositório próprio do workspace", name)
		}
	}
	for _, entry := range existing {
		if entry.Name == name {
			return fmt.Errorf("%q já está registrado", name)
		}
	}
	return nil
}

// AddRepositoryRequest é o pedido de registro de um repositório adicional.
type AddRepositoryRequest struct {
	Name      string
	PathInput string
	Replace   bool
}

// AddRepositoryResult descreve o desfecho comunicado ao usuário.
type AddRepositoryResult struct {
	ProjectName  string
	Name         string
	PreviousPath string
	NewPath      string
	Changed      bool
}

// AddRepository registra um repositório Git local adicional no manifesto. Reaproveita integralmente
// o pipeline de validação de Link: localizar workspace, ler manifesto, resolver caminho, validar o
// repositório e verificar a separação contra todo o conjunto já associado ao workspace.
func AddRepository(start string, request AddRepositoryRequest, inspect LinkGitInspect) (AddRepositoryResult, error) {
	root, manifestPath, err := locateLinkWorkspace(start)
	if err != nil {
		return AddRepositoryResult{}, err
	}
	data, err := readLinkManifest(manifestPath)
	if err != nil {
		return AddRepositoryResult{}, linkFailure("manifest-invalid", "manifesto ausente ou inválido", manifestPath, "corrija ou restaure knowledge/cerne.json")
	}
	if data.VersionErr != nil {
		return AddRepositoryResult{}, linkFailure("manifest-version-unsupported", "versão do manifesto não suportada", manifestPath, "use version como inteiro JSON 1 ou remova o campo")
	}
	if data.WorkflowErr != nil {
		return AddRepositoryResult{}, linkFailure("workflow-invalid", "workflow inválido no manifesto", manifestPath, "corrija o objeto workflow.provider")
	}
	if inspect == nil {
		return AddRepositoryResult{}, linkFailure("git-unavailable", "Git indisponível", "", "instale o Git e disponibilize-o no PATH")
	}

	existingIndex := indexOfRepository(data.Repositories, request.Name)
	if err := validateNewRepositoryName(request.Name, data.Repositories, existingIndex); err != nil {
		return AddRepositoryResult{}, err
	}

	knowledge := filepath.Join(root, "knowledge")
	if err := regularDir(knowledge); err != nil {
		return AddRepositoryResult{}, linkFailure("knowledge-missing", "repositório de conhecimento não encontrado", knowledge, "restaure o diretório knowledge")
	}
	candidate, err := resolveLinkPath(start, request.PathInput)
	if err != nil {
		return AddRepositoryResult{}, err
	}
	candidateFacts, err := validLinkRepository(inspect, candidate)
	if err != nil {
		return AddRepositoryResult{}, err
	}
	knowledgeFacts, err := validLinkRepository(inspect, knowledge)
	if err != nil {
		return AddRepositoryResult{}, linkFailure("knowledge-invalid", "repositório knowledge inválido", knowledge, "inicialize ou restaure knowledge como repositório Git local")
	}

	result := AddRepositoryResult{ProjectName: data.Name, Name: request.Name}
	if existingIndex >= 0 {
		result.PreviousPath = data.Repositories[existingIndex].Path
		previous, err := validateSourcePath(knowledge, result.PreviousPath)
		if err == nil && linkSameSource(inspect, previous, candidate) {
			result.NewPath = result.PreviousPath
			return result, nil
		}
		if !request.Replace {
			return AddRepositoryResult{}, linkFailure("repository-already-registered", "outro caminho já está registrado para "+request.Name, previous, "execute novamente com --replace para substituir apenas a referência do manifesto")
		}
	}

	// O próprio nome é excluído do conjunto: em --replace a entrada antiga está sendo substituída.
	existing := workspaceNamedFacts(inspect, knowledge, knowledgeFacts, data, request.Name)
	if err := validateRepositorySeparation(candidate, candidateFacts, existing); err != nil {
		return AddRepositoryResult{}, err
	}

	newPath := manifestLinkSource(knowledge, candidate)
	entries := append([]RepositoryEntry(nil), data.Repositories...)
	if existingIndex >= 0 {
		entries[existingIndex].Path = newPath
	} else {
		entries = append(entries, RepositoryEntry{Name: request.Name, Path: newPath})
	}
	if err := writeRepositories(manifestPath, data, entries); err != nil {
		return AddRepositoryResult{}, err
	}
	result.NewPath, result.Changed = newPath, true
	return result, nil
}

func indexOfRepository(entries []RepositoryEntry, name string) int {
	for index, entry := range entries {
		if entry.Name == name {
			return index
		}
	}
	return -1
}

// validateNewRepositoryName aplica as regras de nome ignorando a própria entrada quando ela já
// existe, para que --replace sobre um nome registrado não seja recusado como duplicata.
func validateNewRepositoryName(name string, entries []RepositoryEntry, skipIndex int) error {
	others := make([]RepositoryEntry, 0, len(entries))
	for index, entry := range entries {
		if index != skipIndex {
			others = append(others, entry)
		}
	}
	if err := ValidateName(name); err != nil {
		return linkFailure("repository-name-invalid", "nome de repositório inválido", "", "use de 1 a 255 caracteres ASCII, comece por letra ou número e evite espaços e separadores de caminho")
	}
	for _, reserved := range reservedRepositoryNames {
		if name == reserved {
			return linkFailure("repository-name-reserved", "nome reservado para um repositório próprio do workspace", "", "escolha um nome diferente de source e knowledge")
		}
	}
	for _, entry := range others {
		if entry.Name == name {
			return linkFailure("repository-name-duplicate", "nome já registrado no workspace", "", "escolha outro nome ou remova o registro existente")
		}
	}
	return nil
}

// writeRepositories grava a lista no manifesto preservando todos os demais campos, inclusive os
// desconhecidos, e mantém a gravação atômica.
func writeRepositories(manifestPath string, data linkManifest, entries []RepositoryEntry) error {
	if len(entries) == 0 {
		delete(data.raw, "repositories")
	} else {
		encoded, err := json.Marshal(entries)
		if err != nil {
			return linkFailure("manifest-update-unsafe", "manifesto não pode ser atualizado com segurança", manifestPath, "verifique o manifesto e tente novamente")
		}
		data.raw["repositories"] = encoded
	}
	content, err := json.MarshalIndent(data.raw, "", "  ")
	if err != nil {
		return linkFailure("manifest-update-unsafe", "manifesto não pode ser atualizado com segurança", manifestPath, "verifique o manifesto e tente novamente")
	}
	if err := writeManifestAtomically(manifestPath, append(content, '\n')); err != nil {
		return linkFailure("manifest-update-failed", "manifesto não pode ser atualizado com segurança", manifestPath, "verifique permissões e tente novamente")
	}
	return nil
}

// RemoveRepositoryResult descreve a remoção comunicada ao usuário.
type RemoveRepositoryResult struct {
	ProjectName string
	Name        string
	RemovedPath string
}

// RemoveRepository desvincula um repositório adicional do manifesto. Não inspeciona Git e não toca
// o repositório no disco: é o caminho suportado para limpar um vínculo quebrado sem editar o
// manifesto à mão.
func RemoveRepository(start string, name string) (RemoveRepositoryResult, error) {
	_, manifestPath, err := locateLinkWorkspace(start)
	if err != nil {
		return RemoveRepositoryResult{}, err
	}
	data, err := readLinkManifest(manifestPath)
	if err != nil {
		return RemoveRepositoryResult{}, linkFailure("manifest-invalid", "manifesto ausente ou inválido", manifestPath, "corrija ou restaure knowledge/cerne.json")
	}
	if data.VersionErr != nil {
		return RemoveRepositoryResult{}, linkFailure("manifest-version-unsupported", "versão do manifesto não suportada", manifestPath, "use version como inteiro JSON 1 ou remova o campo")
	}
	for _, reserved := range reservedRepositoryNames {
		if name == reserved {
			return RemoveRepositoryResult{}, linkFailure("repository-not-removable", "repositório próprio do workspace não pode ser desvinculado", name, "use cerne link para trocar o source; knowledge é parte do workspace")
		}
	}
	index := indexOfRepository(data.Repositories, name)
	if index < 0 {
		return RemoveRepositoryResult{}, linkFailure("repository-not-registered", "repositório não registrado no workspace", name, "use um nome listado por cerne status")
	}
	removed := data.Repositories[index]
	entries := append(append([]RepositoryEntry(nil), data.Repositories[:index]...), data.Repositories[index+1:]...)
	if err := writeRepositories(manifestPath, data, entries); err != nil {
		return RemoveRepositoryResult{}, err
	}
	return RemoveRepositoryResult{ProjectName: data.Name, Name: name, RemovedPath: removed.Path}, nil
}
