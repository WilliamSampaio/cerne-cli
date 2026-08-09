#!/bin/sh
set -eu

repo_url="${CERNE_INSTALL_REPO_URL:-https://github.com/WilliamSampaio/cerne-cli/releases}"
version=""
agent=""

usage() {
	cat <<'EOF'
Usage:
  install.sh [--version <version>] [--agent <codex|claude>]
  install.sh --help

Installs cerne to ~/.local/bin/cerne without sudo.
EOF
}

fail() {
	printf 'error: %s\n' "$1" >&2
	exit 1
}

need_value() {
	test "$#" -gt 1 || fail "$1 requires a value"
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
			case "$version" in ""|*/*|*\\*) fail "invalid version: $version" ;; esac
			shift 2
			;;
		--agent)
			need_value "$@"
			agent=$2
			case "$agent" in
				codex|claude) ;;
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

if test -e "$target" && ! test -f "$target"; then
	fail "$target exists and is not a regular file"
fi
if test -L "$target"; then
	fail "$target is a symbolic link"
fi

tmp=${TMPDIR:-/tmp}/cerne-install.$$
cleanup() {
	case "${tmp:-}" in
		""|"/"|"/tmp") ;;
		*) test -d "$tmp" && rm -rf "$tmp" ;;
	esac
}
trap cleanup EXIT HUP INT TERM
mkdir -p "$tmp" "$bin_dir"

download() {
	url=$1
	out=$2
	case "$url" in
		file://*)
			test "${CERNE_INSTALL_REPO_URL:-}" != "" || fail "file downloads are test-only"
			cp "${url#file://}" "$out"
			return
			;;
	esac
	if command -v curl >/dev/null 2>&1; then
		curl --proto '=https' --tlsv1.2 -fsSL "$url" -o "$out"
	elif command -v wget >/dev/null 2>&1; then
		wget -q -O "$out" "$url"
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
	archive=$(awk -v os="$os" -v arch="$arch" '$2 ~ "^cerne_.*_" os "_" arch "\\.tar\\.gz$" { print $2 }' "$checksums")
	count=$(printf '%s\n' "$archive" | sed '/^$/d' | wc -l | tr -d ' ')
	test "$count" = "1" || fail "compatible release artifact was not found"
else
	count=$(awk -v file="$archive" '$2 == file { n++ } END { print n+0 }' "$checksums")
	test "$count" = "1" || fail "checksum entry was not found for $archive"
fi

expected=$(awk -v file="$archive" '$2 == file { print $1 }' "$checksums")
printf '%s\n' "$expected" | grep '^[0-9a-fA-F]\{64\}$' >/dev/null || fail "invalid checksum for $archive"

tarball=$tmp/$archive
download "$asset_base/$archive" "$tarball" || fail "failed to download $archive"
actual=$(sha256_file "$tarball")
test "$actual" = "$expected" || fail "checksum mismatch for $archive"

mkdir -p "$tmp/extract"
test "$(tar -tzf "$tarball")" = "cerne" || fail "archive must contain only cerne at root"
tar -xzf "$tarball" -C "$tmp/extract" || fail "failed to extract $archive"
test -f "$tmp/extract/cerne" || fail "archive does not contain cerne"
test -x "$tmp/extract/cerne" || chmod +x "$tmp/extract/cerne"
mv "$tmp/extract/cerne" "$target" || fail "failed to install cerne"

installed=$("$target" --version 2>/dev/null || true)
test -n "$installed" || fail "installed binary did not report a version"
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
