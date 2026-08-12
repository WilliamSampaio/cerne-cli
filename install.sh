#!/bin/sh
set -eu

repo_url="${CERNE_INSTALL_REPO_URL:-https://github.com/WilliamSampaio/cerne-cli/releases}"
version=""
agent=""
allow_file_downloads=0

usage() {
	cat <<'EOF'
Usage:
  install.sh [--version <version>] [--agent <codex|claude|gemini>]
  install.sh --help

Installs cerne to ~/.local/bin/cerne without sudo.
EOF
}

fail() {
	printf 'error: %s\n' "$1" >&2
	exit 1
}

ensure_safe_dir_under_home() {
	current=$home
	if test -L "$current"; then
		fail "$current is a symbolic link"
	fi
	if test -e "$current"; then
		test -d "$current" || fail "$current exists and is not a directory"
	else
		mkdir "$current" || fail "failed to create $current"
	fi
	old_ifs=$IFS
	IFS=/
	set -- $1
	IFS=$old_ifs
	for part do
		test -n "$part" || continue
		current=$current/$part
		if test -L "$current"; then
			fail "$current is a symbolic link"
		fi
		if test -e "$current"; then
			test -d "$current" || fail "$current exists and is not a directory"
		else
			mkdir "$current" || fail "failed to create $current"
			test ! -L "$current" || fail "$current is a symbolic link"
		fi
	done
}

need_value() {
	test "$#" -gt 1 || fail "$1 requires a value"
}

is_release_version() {
	printf '%s\n' "$1" | awk '
	function valid_identifiers(value, reject_numeric_leading_zero, n, parts, i) {
		n = split(value, parts, ".")
		for (i = 1; i <= n; i++) {
			if (parts[i] !~ /^[0-9A-Za-z-]+$/) return 0
			if (reject_numeric_leading_zero && parts[i] ~ /^[0-9]+$/ && parts[i] ~ /^0[0-9]/) return 0
		}
		return 1
	}
	{
		value = $0
		plus = index(value, "+")
		if (plus) {
			build = substr(value, plus + 1)
			value = substr(value, 1, plus - 1)
			if (!valid_identifiers(build, 0)) exit 1
			if (index(value, "+")) exit 1
		}
		dash = index(value, "-")
		if (dash) {
			pre = substr(value, dash + 1)
			value = substr(value, 1, dash - 1)
			if (!valid_identifiers(pre, 1)) exit 1
		}
		if (value ~ /^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$/) ok = 1
	}
	END { exit ok ? 0 : 1 }'
}

while test "$#" -gt 0; do
	case "$1" in
		--help)
			usage
			exit 0
			;;
		--version)
			need_value "$@"
			version=$2
			is_release_version "$version" || fail "invalid version: $version"
			shift 2
			;;
		--agent)
			need_value "$@"
			agent=$2
			case "$agent" in
				codex|claude|gemini) ;;
				*) fail "unsupported agent: $agent" ;;
			esac
			shift 2
			;;
		*)
			fail "unknown argument: $1"
			;;
	esac
done

os=${CERNE_INSTALL_OS:-$(uname -s | tr '[:upper:]' '[:lower:]')}
arch=${CERNE_INSTALL_ARCH:-$(uname -m)}
case "$os" in
	linux|darwin) ;;
	*) fail "unsupported operating system: $os" ;;
esac
case "$arch" in
	x86_64|amd64) arch=amd64 ;;
	aarch64|arm64) arch=arm64 ;;
	*) fail "unsupported architecture: $arch" ;;
esac

home=${HOME:-}
test -n "$home" || fail "HOME is not set"
bin_dir=$home/.local/bin
target=$bin_dir/cerne
case "$target" in
	"$home"/*) ;;
	*) fail "install destination must be inside HOME" ;;
esac

if test -L "$target"; then
	fail "$target is a symbolic link"
fi
if test -e "$target" && ! test -f "$target"; then
	fail "$target exists and is not a regular file"
fi

old_umask=$(umask)
umask 077
tmp=$(mktemp -d "${TMPDIR:-/tmp}/cerne-install.XXXXXX") || fail "failed to create temporary directory"
umask "$old_umask"
tmp_marker=$tmp/.cerne-owned
staged=""
: >"$tmp_marker"
cleanup() {
	case "${staged:-}" in
		"$bin_dir"/.cerne.*) test -f "$staged" && rm -f "$staged" ;;
	esac
	case "${tmp:-}" in
		""|"/"|"/tmp") ;;
		*) test -d "$tmp" && test ! -L "$tmp" && test -f "$tmp_marker" && rm -rf "$tmp" ;;
	esac
}
trap cleanup EXIT HUP INT TERM
ensure_safe_dir_under_home ".local/bin"

case "$repo_url" in
	https://*) ;;
	file://*) test "${CERNE_INSTALL_ALLOW_FILE_URLS:-}" = "1" || fail "file downloads require CERNE_INSTALL_ALLOW_FILE_URLS=1"; allow_file_downloads=1 ;;
	*) fail "release URL must use https" ;;
esac

download() {
	url=$1
	out=$2
	case "$url" in
		file://*)
			test "$allow_file_downloads" = "1" || fail "file downloads are test-only"
			cp "${url#file://}" "$out"
			return
			;;
		https://*) ;;
		*) fail "download URL must use https" ;;
	esac
	if command -v curl >/dev/null 2>&1; then
		curl --proto '=https' --tlsv1.2 -fsSL "$url" -o "$out"
	elif command -v wget >/dev/null 2>&1; then
		wget --help 2>/dev/null | grep -- '--https-only' >/dev/null || fail "curl or wget with HTTPS-only support is required"
		wget --https-only -q -O "$out" "$url"
	else
		fail "curl or wget is required"
	fi
}

sha256_file() {
	file=$1
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$file" | awk '{print $1}'
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$file" | awk '{print $1}'
	else
		fail "sha256sum or shasum is required"
	fi
}

binary_reports_version() {
	binary=$1
	expected_version=$2
	output=$("$binary" --version 2>/dev/null || true)
	test "$output" = "cerne $expected_version"
}

if test -n "$version"; then
	asset_base=$repo_url/download/$version
	archive=cerne_${version}_${os}_${arch}.tar.gz
else
	asset_base=$repo_url/latest/download
	archive=""
fi

checksums=$tmp/checksums.txt
download "$asset_base/checksums.txt" "$checksums" || fail "failed to download checksums"

if test -z "$archive"; then
	archive=$(awk -v os="$os" -v arch="$arch" '$2 ~ "^cerne_v[0-9][0-9]*\\.[0-9][0-9]*\\.[0-9][0-9]*(-[0-9A-Za-z][0-9A-Za-z.-]*)?(\\+[0-9A-Za-z][0-9A-Za-z.-]*)?_" os "_" arch "\\.tar\\.gz$" { print $2 }' "$checksums")
	count=$(printf '%s\n' "$archive" | sed '/^$/d' | wc -l | tr -d ' ')
	test "$count" = "1" || fail "compatible release artifact was not found"
else
	count=$(awk -v file="$archive" '$2 == file { n++ } END { print n+0 }' "$checksums")
	test "$count" = "1" || fail "checksum entry was not found for $archive"
fi
archive_version=${archive#cerne_}
archive_version=${archive_version%_${os}_${arch}.tar.gz}
is_release_version "$archive_version" || fail "invalid archive name: $archive"

expected=$(awk -v file="$archive" '$2 == file { print $1 }' "$checksums")
printf '%s\n' "$expected" | grep '^[0-9a-fA-F]\{64\}$' >/dev/null || fail "invalid checksum for $archive"

tarball=$tmp/$archive
download "$asset_base/$archive" "$tarball" || fail "failed to download $archive"
actual=$(sha256_file "$tarball")
test "$actual" = "$expected" || fail "checksum mismatch for $archive"

mkdir -p "$tmp/extract"
test "$(tar -tzf "$tarball")" = "cerne" || fail "archive must contain only cerne at root"
test "$(tar -tzvf "$tarball" | awk 'NR == 1 { print substr($1, 1, 1) }')" = "-" || fail "archive member cerne must be a regular file"
tar -xzf "$tarball" -C "$tmp/extract" || fail "failed to extract $archive"
test -f "$tmp/extract/cerne" || fail "archive does not contain cerne"
test ! -L "$tmp/extract/cerne" || fail "archive member cerne must not be a symbolic link"
test -x "$tmp/extract/cerne" || chmod +x "$tmp/extract/cerne"
binary_reports_version "$tmp/extract/cerne" "$archive_version" || fail "downloaded binary did not report $archive_version"
staged=$(mktemp "$bin_dir/.cerne.XXXXXX") || fail "failed to stage cerne"
cp "$tmp/extract/cerne" "$staged" || fail "failed to stage cerne"
chmod 755 "$staged" || fail "failed to prepare staged cerne"
binary_reports_version "$staged" "$archive_version" || fail "staged binary did not report $archive_version"
ensure_safe_dir_under_home ".local/bin"
mv "$staged" "$target" || fail "failed to install cerne"
staged=""

installed=$("$target" --version 2>/dev/null || true)
test "$installed" = "cerne $archive_version" || fail "installed binary did not report $archive_version"
printf 'installed: %s\n' "$installed"
printf 'path: %s\n' "$target"
case ":$PATH:" in
	*":$bin_dir:"*) ;;
	*) printf 'add to PATH: export PATH="$HOME/.local/bin:$PATH"\n' ;;
esac

if test -n "$agent"; then
	if "$target" skill install "$agent"; then
		printf 'skill installed for: %s\n' "$agent"
	else
		fail "cerne was installed, but skill installation failed for $agent"
	fi
fi
