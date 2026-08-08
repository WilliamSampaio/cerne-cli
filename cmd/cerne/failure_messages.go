package main

import (
	"fmt"
	"io"

	"github.com/WilliamSampaio/cerne-cli/internal/localization"
)

type failureMessage struct {
	Cause      string
	Correction string
}

func localizedFailure(messages localizer, domain, code, fallbackCause, fallbackCorrection string) failureMessage {
	if translated, ok := failureCatalogs[messages.language][domain+"."+code]; ok {
		return translated
	}
	if messages.language == localization.PortugueseBrazil {
		return failureMessage{Cause: fallbackCause, Correction: fallbackCorrection}
	}
	return failureMessage{Cause: messages.text("failure.operational"), Correction: messages.text("failure.check-and-retry")}
}

func localizedFailureCause(messages localizer, domain, code, fallback string) string {
	return localizedFailure(messages, domain, code, fallback, "").Cause
}

func renderFailure(output io.Writer, messages localizer, domain, code, fallbackCause, path, fallbackCorrection string) {
	failure := localizedFailure(messages, domain, code, fallbackCause, fallbackCorrection)
	if path == "" {
		fmt.Fprint(output, messages.text("failure.cause", failure.Cause))
	} else {
		fmt.Fprint(output, messages.text("failure.cause.path", failure.Cause, path))
	}
	if failure.Correction != "" {
		fmt.Fprint(output, messages.text("failure.correction", failure.Correction))
	}
}

var failureCatalogs = map[localization.Language]map[string]failureMessage{
	localization.PortugueseBrazil: portugueseFailureMessages,
	localization.English:          englishFailureMessages,
}
