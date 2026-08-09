package main

import (
	"fmt"

	"github.com/WilliamSampaio/cerne-cli/internal/localization"
)

type messageID string

const (
	messageGlobalHelp          messageID = "help.global"
	messageConfigHelp          messageID = "help.config"
	messageSkillHelp           messageID = "help.skill"
	messageContextHelp         messageID = "help.context"
	messageRestoreHelp         messageID = "help.restore"
	messageInitHelp            messageID = "help.init"
	messageWorkflowHelp        messageID = "help.workflow"
	messageDoctorHelp          messageID = "help.doctor"
	messageStatusHelp          messageID = "help.status"
	messageLinkHelp            messageID = "help.link"
	messageGitHelp             messageID = "help.git"
	messageInvalidLanguage     messageID = "error.invalid-language"
	messageInvalidGlobalOption messageID = "error.invalid-global-option"
	messageConfigUsage         messageID = "error.config-usage"
	messageConfigSet           messageID = "config.set"
	messageConfigGet           messageID = "config.get"
	messageConfigGetUnset      messageID = "config.get-unset"
	messageConfigUnset         messageID = "config.unset"
	messageConfigUnsafe        messageID = "error.config-unsafe"
	messageConfigRead          messageID = "error.config-read"
	messageConfigInvalid       messageID = "error.config-invalid"
	messageConfigWrite         messageID = "error.config-write"
	messageHomeUnavailable     messageID = "error.home-unavailable"
)

type localizer struct {
	language localization.Language
}

func (localizer localizer) text(id messageID, args ...any) string {
	template, ok := localizer.find(id)
	if !ok {
		return fmt.Sprintf("[missing message: %s]", id)
	}
	if len(args) == 0 {
		return template
	}
	return fmt.Sprintf(template, args...)
}

func (localizer localizer) find(id messageID) (string, bool) {
	template, ok := messageCatalogs[localizer.language][id]
	return template, ok
}

var messageCatalogs = map[localization.Language]map[messageID]string{
	localization.English:          englishMessages,
	localization.PortugueseBrazil: portugueseBrazilMessages,
}
