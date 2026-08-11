#!/bin/sh
set -eu

tag=${1:-}
semver='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-((0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*)(\.(0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*))*))?(\+([0-9A-Za-z-]+)(\.[0-9A-Za-z-]+)*)?$'

printf '%s\n' "$tag" | grep -Eq "$semver" || {
	printf 'invalid release tag: %s\n' "$tag" >&2
	exit 1
}
