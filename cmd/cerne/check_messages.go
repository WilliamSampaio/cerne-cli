package main

import (
	"github.com/WilliamSampaio/cerne-cli/internal/localization"
	"github.com/WilliamSampaio/cerne-cli/internal/workspace"
)

type checkMessage struct {
	Detail     string
	Correction string
}

func localizedCheck(messages localizer, check workspace.CheckResult) (string, string, string) {
	if messages.language == localization.PortugueseBrazil {
		return check.Label, check.Detail, check.Correction
	}
	label := englishCheckLabels[check.ID]
	if label == "" {
		label = check.ID
	}
	translated, ok := englishCheckMessages[check.ID+"."+check.Code]
	if !ok {
		return label, messages.text("failure.operational"), messages.text("failure.check-and-retry")
	}
	return label, translated.Detail, translated.Correction
}

var englishCheckLabels = map[string]string{
	"manifest":              "Manifest",
	"knowledge-repository":  "Knowledge repository",
	"source-repository":     "Source repository",
	"git-independence":      "Git independence",
	"versioning-isolation":  "Versioning isolation",
	"manifest-paths":        "Manifest paths",
	"knowledge-directories": "Required directories",
	"workflow":              "Workflow",
	"git-available":         "Git",
	"permissions":           "Permissions",
	"manifest-version":      "Manifest version",
}

var englishCheckMessages = map[string]checkMessage{
	"manifest.invalid":                            {"invalid or unreadable", "repair knowledge/cerne.json"},
	"manifest.name-mismatch":                      {"valid name differs from root name", "align the manifest or rename the workspace"},
	"manifest.readable":                           {Detail: "readable"},
	"knowledge-repository.missing":                {"not found as a regular directory", "create or restore the expected directory"},
	"knowledge-repository.found":                  {Detail: "found"},
	"source-repository.missing":                   {"not found as a regular directory", "create or restore the expected directory"},
	"source-repository.found":                     {Detail: "found"},
	"git-independence.git-unavailable":            {"Git is unavailable", "install Git and make it available in PATH"},
	"git-independence.roots-unconfirmed":          {"independent Git roots were not confirmed", "initialize knowledge and source as independent Git repositories"},
	"git-independence.history-shared":             {"Git history is shared", "use independent Git repositories"},
	"git-independence.independent":                {Detail: "roots and histories are distinct"},
	"versioning-isolation.git-unavailable":        {"Git is unavailable", "install Git and make it available in PATH"},
	"versioning-isolation.repository-contained":   {"one repository contains the other", "keep knowledge and source as sibling directories"},
	"versioning-isolation.boundaries-unconfirmed": {"Git boundaries were not confirmed", "repair the local Git repositories"},
	"versioning-isolation.worktree-contained":     {"one worktree contains the other", "separate the Git repositories"},
	"versioning-isolation.isolated":               {Detail: "neither repository contains the other"},
	"manifest-paths.manifest-invalid":             {"invalid manifest prevents path resolution", "repair knowledge/cerne.json"},
	"manifest-paths.source-invalid":               {"invalid source", "configure an existing and safe source path"},
	"manifest-paths.source-missing":               {"source is not a regular directory", "restore the source directory"},
	"manifest-paths.valid":                        {Detail: "valid"},
	"knowledge-directories.knowledge-unavailable": {"knowledge is unavailable", "restore the knowledge repository"},
	"knowledge-directories.required-missing":      {"a required directory is missing or invalid", "restore product, specs, decisions, policies, and runs"},
	"knowledge-directories.found-without-specs":   {Detail: "product, decisions, policies, and runs found"},
	"knowledge-directories.found":                 {Detail: "product, specs, decisions, policies, and runs found"},
	"workflow.config-invalid":                     {"invalid configuration", "repair workflow.provider in the manifest"},
	"workflow.resolve-failed":                     {"provider could not be resolved", "run doctor with an updated Cerne version"},
	"workflow.unknown-provider":                   {"unknown provider", "use speckit or openspec in a new workspace"},
	"workflow.definition-invalid":                 {"invalid definition", "update Cerne"},
	"workflow.layout-invalid":                     {"invalid or partial layout", "repair the provider layout before setup"},
	"workflow.configured-pending":                 {"configured but pending", "run cerne workflow setup"},
	"workflow.executable-missing":                 {"executable missing; workflow pending", "install the provider and run cerne workflow setup"},
	"workflow.specs-missing":                      {"canonical specifications directory missing", "restore the provider layout"},
	"workflow.materialized-executable-missing":    {"materialized; executable missing", "install the provider to run workflow commands"},
	"workflow.ready":                              {Detail: "materialized and available"},
	"git-available.unavailable":                   {"unavailable", "install Git and make it available in PATH"},
	"git-available.available":                     {Detail: "available"},
	"permissions.unconfirmed":                     {"could not confirm permissions", "verify read and write access manually"},
	"permissions.denied":                          {"read or write access denied", "adjust workspace permissions"},
	"permissions.read-write-unconfirmed":          {"could not confirm read and write access", "verify workspace permissions"},
	"permissions.confirmed":                       {Detail: "read and write access confirmed"},
	"manifest-version.unsupported":                {"unsupported version", "use version as JSON integer 1 or remove the field"},
	"manifest-version.explicit-v1":                {Detail: "version 1 supported"},
	"manifest-version.implicit-v1":                {Detail: "implicit version 1 supported"},
}
