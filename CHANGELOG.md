# Changelog

All notable changes to Cerne are documented in this file. This project follows
[Semantic Versioning](https://semver.org/).

## 0.1.0 - 2026-08-03

### Added

- `cerne init` to create workspaces with independent knowledge and source repositories.
- `cerne doctor` to validate workspace structure, safety, permissions, and manifest compatibility.
- `cerne status` to report local Git state for both repositories without modifying them.
- `cerne link` to atomically link an existing local Git worktree as source.
- Global `--help` and `--version` options.
- English, Brazilian Portuguese, and Spanish documentation.
- Automated tests on Linux, Windows, and macOS.

### Security

- Git inspection disables prompts and optional locks and sanitizes redirecting `GIT_*` variables.
- Workspace operations reject unsafe repository overlap and preserve existing files by default.
- Build baseline updated to Go 1.26.5 and `golang.org/x/sys` v0.47.0.
