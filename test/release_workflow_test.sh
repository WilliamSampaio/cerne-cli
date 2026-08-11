#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
validator=$root/scripts/validate-release-tag.sh
workflow=$root/.github/workflows/release.yml

fail() {
	printf 'release workflow test: %s\n' "$1" >&2
	exit 1
}

test -f "$validator" || fail "missing release tag validator"

for tag in v0.1.0 v1.26.5 v2.0.0-rc.1 v2.0.0+build.7 v2.0.0-rc.1+build.7; do
	sh "$validator" "$tag" >/dev/null || fail "valid SemVer tag rejected: $tag"
done

for tag in v1 v1.2 1.2.3 v01.2.3 v1.02.3 v1.2.03 v1.2.3-01 v1.2.3- v1.2.3+; do
	if sh "$validator" "$tag" >/dev/null 2>&1; then
		fail "invalid SemVer tag accepted: $tag"
	fi
done

grep -F 'validate-tag:' "$workflow" >/dev/null || fail "missing tag validation job"
grep -F 'os: [ubuntu-latest, windows-latest, macos-latest]' "$workflow" >/dev/null || fail "missing complete OS test matrix"
grep -F 'needs: validate-tag' "$workflow" >/dev/null || fail "test matrix is not gated by tag validation"
grep -F 'needs: [validate-tag, test]' "$workflow" >/dev/null || fail "release build is not gated by tag validation and tests"
grep -F 'run: go test -count=1 ./...' "$workflow" >/dev/null || fail "release matrix does not run the Go suite"
