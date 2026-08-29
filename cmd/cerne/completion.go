package main

import (
	"fmt"
	"io"
	"strings"
)

// completionItem is a single completion candidate: the literal word plus the message key for its
// localized one-line description (used only by the zsh script — bash has no description support).
type completionItem struct {
	name string
	desc messageID
}

var topLevelCompletionItems = []completionItem{
	{"init", "completion.desc.init"},
	{"restore", "completion.desc.restore"},
	{"doctor", "completion.desc.doctor"},
	{"status", "completion.desc.status"},
	{"link", "completion.desc.link"},
	{"workflow", "completion.desc.workflow"},
	{"context", "completion.desc.context"},
	{"skill", "completion.desc.skill"},
	{"git", "completion.desc.git"},
	{"config", "completion.desc.config"},
	{"completion", "completion.desc.completion"},
}

// Second-level completion is limited to subcommands with a fixed, closed vocabulary at that
// position. Free-form arguments (paths, project names, runtimes, languages) remain out of scope.
var secondLevelCompletionItems = []struct {
	parent string
	items  []completionItem
}{
	{"config", []completionItem{
		{"set", "completion.desc.config.set"},
		{"get", "completion.desc.config.get"},
		{"unset", "completion.desc.config.unset"},
	}},
	{"skill", []completionItem{{"install", "completion.desc.skill.install"}}},
	{"workflow", []completionItem{{"setup", "completion.desc.workflow.setup"}}},
	{"git", []completionItem{{"inspect", "completion.desc.git.inspect"}}},
}

// Third level: the only closed-vocabulary word after "config set|get|unset" is the literal
// argument "language" (config manages no other key).
var configLanguageCompletionItems = []completionItem{{"language", "completion.desc.config.language"}}

// Third level after "skill install", and the value accepted by "--runtime" wherever it appears
// (workflow setup, git inspect).
var agentCompletionItems = []completionItem{
	{"codex", "completion.desc.agent.codex"},
	{"claude", "completion.desc.agent.claude"},
	{"gemini", "completion.desc.agent.gemini"},
}

// Fourth level: the skill name accepted by "skill install <agent>", which depends on the agent.
var skillNameCompletionItems = []completionItem{
	{"cerne-context", "completion.desc.skillname.cerne-context"},
	{"cerne-product-discovery", "completion.desc.skillname.cerne-product-discovery"},
	{"cerne-git-workflow", "completion.desc.skillname.cerne-git-workflow"},
}

var geminiSkillNameCompletionItems = []completionItem{
	{"cerne-git-workflow", "completion.desc.skillname.cerne-git-workflow"},
}

func completionNames(items []completionItem) string {
	names := make([]string, len(items))
	for i, item := range items {
		names[i] = item.name
	}
	return strings.Join(names, " ")
}

var completionSubcommands = completionNames(topLevelCompletionItems)

const bashCompletionTemplate = `_cerne_completions() {
  if [ "$COMP_CWORD" -eq 1 ]; then
    COMPREPLY=( $(compgen -W "%s" -- "${COMP_WORDS[1]}") )
    return
  fi
  if [ "$COMP_CWORD" -eq 2 ]; then
    case "${COMP_WORDS[1]}" in
%s
    esac
    return
  fi
  if [ "$COMP_CWORD" -eq 3 ]; then
    case "${COMP_WORDS[1]} ${COMP_WORDS[2]}" in
      "config set"|"config get"|"config unset") COMPREPLY=( $(compgen -W "%s" -- "${COMP_WORDS[3]}") ) ;;
      "skill install") COMPREPLY=( $(compgen -W "%s" -- "${COMP_WORDS[3]}") ) ;;
    esac
    return
  fi
  if [ "$COMP_CWORD" -eq 4 ]; then
    if [ "${COMP_WORDS[3]}" = "--runtime" ] && { [ "${COMP_WORDS[1]}" = "workflow" ] || [ "${COMP_WORDS[1]}" = "git" ]; }; then
      COMPREPLY=( $(compgen -W "%s" -- "${COMP_WORDS[4]}") )
    elif [ "${COMP_WORDS[1]}" = "skill" ] && [ "${COMP_WORDS[2]}" = "install" ]; then
      case "${COMP_WORDS[3]}" in
        gemini) COMPREPLY=( $(compgen -W "%s" -- "${COMP_WORDS[4]}") ) ;;
        codex|claude) COMPREPLY=( $(compgen -W "%s" -- "${COMP_WORDS[4]}") ) ;;
      esac
    fi
  fi
}
complete -F _cerne_completions cerne
`

func bashCompletionScriptText() string {
	var cases strings.Builder
	for _, entry := range secondLevelCompletionItems {
		fmt.Fprintf(&cases, "      %s) COMPREPLY=( $(compgen -W \"%s\" -- \"${COMP_WORDS[2]}\") ) ;;\n",
			entry.parent, completionNames(entry.items))
	}
	agentNames := completionNames(agentCompletionItems)
	return fmt.Sprintf(bashCompletionTemplate,
		completionSubcommands,
		strings.TrimSuffix(cases.String(), "\n"),
		completionNames(configLanguageCompletionItems),
		agentNames,
		agentNames,
		completionNames(geminiSkillNameCompletionItems),
		completionNames(skillNameCompletionItems),
	)
}

const zshCompletionTemplate = `#compdef cerne

if ! (( $+functions[compdef] )); then
  autoload -Uz compinit && compinit -u
fi

_cerne() {
  if (( CURRENT == 2 )); then
    local -a subcommands
    subcommands=(
%s
    )
    _describe 'command' subcommands
  elif (( CURRENT == 3 )); then
    local -a actions
    case ${words[2]} in
%s
    esac
    (( ${#actions} )) && _describe 'action' actions
  elif (( CURRENT == 4 )); then
    local -a values
    case "${words[2]} ${words[3]}" in
      "config set"|"config get"|"config unset") values=(
%s
      ) ;;
      "skill install") values=(
%s
      ) ;;
    esac
    (( ${#values} )) && _describe 'value' values
  elif (( CURRENT == 5 )); then
    local -a values
    if [[ ${words[4]} == --runtime && ( ${words[2]} == workflow || ${words[2]} == git ) ]]; then
      values=(
%s
      )
    elif [[ ${words[2]} == skill && ${words[3]} == install ]]; then
      case ${words[4]} in
        gemini) values=(
%s
        ) ;;
        codex|claude) values=(
%s
        ) ;;
      esac
    fi
    (( ${#values} )) && _describe 'value' values
  fi
}

compdef _cerne cerne
`

// zshDescribeEntry renders a single _describe candidate as 'word:description', the format zsh
// uses to show the description of the item currently selected/highlighted in the completion menu.
func zshDescribeEntry(item completionItem, messages localizer) string {
	return "      '" + item.name + ":" + messages.text(item.desc) + "'"
}

// zshDescribeBlock renders each item on its own line, for embedding inside a "name=( ... )" array.
func zshDescribeBlock(items []completionItem, messages localizer) string {
	lines := make([]string, len(items))
	for i, item := range items {
		lines[i] = zshDescribeEntry(item, messages)
	}
	return strings.Join(lines, "\n")
}

func zshCompletionScriptText(messages localizer) string {
	var subcommands strings.Builder
	for i, item := range topLevelCompletionItems {
		if i > 0 {
			subcommands.WriteString("\n")
		}
		subcommands.WriteString(zshDescribeEntry(item, messages))
	}

	var cases strings.Builder
	for _, entry := range secondLevelCompletionItems {
		fmt.Fprintf(&cases, "      %s) actions=(\n", entry.parent)
		for _, item := range entry.items {
			cases.WriteString(zshDescribeEntry(item, messages))
			cases.WriteString("\n")
		}
		cases.WriteString("      ) ;;\n")
	}

	agentBlock := zshDescribeBlock(agentCompletionItems, messages)
	return fmt.Sprintf(zshCompletionTemplate,
		subcommands.String(),
		strings.TrimSuffix(cases.String(), "\n"),
		zshDescribeBlock(configLanguageCompletionItems, messages),
		agentBlock,
		agentBlock,
		zshDescribeBlock(geminiSkillNameCompletionItems, messages),
		zshDescribeBlock(skillNameCompletionItems, messages),
	)
}

func runCompletion(args []string, stdout, stderr io.Writer, messages localizer) int {
	if len(args) == 1 && args[0] == "--help" {
		fmt.Fprint(stdout, messages.text(messageCompletionHelp))
		return 0
	}
	if len(args) != 1 {
		fmt.Fprint(stderr, messages.text("completion.usage"))
		return 2
	}
	switch args[0] {
	case "bash":
		fmt.Fprint(stdout, bashCompletionScriptText())
		return 0
	case "zsh":
		fmt.Fprint(stdout, zshCompletionScriptText(messages))
		return 0
	default:
		fmt.Fprint(stderr, messages.text("completion.usage"))
		return 2
	}
}
