#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
work=$(mktemp -d "${TMPDIR:-/tmp}/cerne-install-test.XXXXXX")
trap 'rm -rf "$work"' EXIT HUP INT TERM
tmp_root=$work/tmp
mkdir -p "$work/release/download/v1.2.3" "$work/release/latest/download" "$work/src" "$tmp_root"

make_asset() {
	version=$1
	os=$2
	arch=$3
	dir=$4
	skill_status=${5:-0}
	tmp=$work/src/$version-$os-$arch
	mkdir -p "$tmp"
	cat >"$tmp/cerne" <<EOF
#!/bin/sh
if test "\${1:-}" = "--version"; then
  echo "cerne $version"
  exit 0
fi
if test "\${1:-}" = "skill" && test "\${2:-}" = "install"; then
  echo "skill \${3:-}"
  exit $skill_status
fi
exit 2
EOF
	chmod +x "$tmp/cerne"
	archive=cerne_${version}_${os}_${arch}.tar.gz
	( cd "$tmp" && tar -czf "$dir/$archive" cerne )
	sum=$(sha256sum "$dir/$archive" | awk '{print $1}')
	printf '%s  %s\n' "$sum" "$archive" >>"$dir/checksums.txt"
}

make_bad_asset() {
	version=$1
	kind=$2
	dir=$3
	tmp=$work/src/$version-linux-amd64-$kind
	mkdir -p "$tmp"
	case "$kind" in
		directory)
			mkdir "$tmp/cerne"
			;;
		invalid-binary)
			printf '#!/bin/sh\nexit 0\n' >"$tmp/cerne"
			chmod +x "$tmp/cerne"
			;;
		fifo)
			mkfifo "$tmp/cerne"
			;;
		symlink)
			ln -s /tmp/cerne "$tmp/cerne"
			;;
		*)
			echo "unknown bad asset kind: $kind" >&2
			exit 1
			;;
	esac
	archive=cerne_${version}_linux_amd64.tar.gz
	( cd "$tmp" && tar -czf "$dir/$archive" cerne )
	sum=$(sha256sum "$dir/$archive" | awk '{print $1}')
	printf '%s  %s\n' "$sum" "$archive" >>"$dir/checksums.txt"
}

make_asset v1.2.3 linux amd64 "$work/release/download/v1.2.3"
cp "$work/release/download/v1.2.3/"* "$work/release/latest/download/"
mkdir -p "$work/release/download/v1.2.4"
make_asset v1.2.4 linux amd64 "$work/release/download/v1.2.4" 1
mkdir -p "$work/release/download/v1.2.5" "$work/release/download/v1.2.6" "$work/release/download/v1.2.7" "$work/release/download/v1.2.8"
make_bad_asset v1.2.5 invalid-binary "$work/release/download/v1.2.5"
make_bad_asset v1.2.6 symlink "$work/release/download/v1.2.6"
make_bad_asset v1.2.7 directory "$work/release/download/v1.2.7"
make_bad_asset v1.2.8 fifo "$work/release/download/v1.2.8"

run_install() {
	HOME=$work/home PATH=/usr/bin:/bin TMPDIR=$tmp_root CERNE_INSTALL_ALLOW_FILE_URLS=1 CERNE_INSTALL_REPO_URL=file://$work/release CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=amd64 sh "$root/install.sh" "$@"
}

run_install >"$work/install-out"
test -x "$work/home/.local/bin/cerne"
grep 'cerne v1.2.3' "$work/install-out" >/dev/null
grep 'add to PATH' "$work/install-out" >/dev/null

run_install >/dev/null

test ! -e "$work/home/.codex"
test ! -e "$work/home/.claude"

if HOME=$work/missing-home TMPDIR=$tmp_root CERNE_INSTALL_ALLOW_FILE_URLS=1 CERNE_INSTALL_REPO_URL=file://$work/release CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=amd64 sh "$root/install.sh" --version v9.9.9 >/dev/null 2>/dev/null; then
	echo "expected missing version failure" >&2
	exit 1
fi
test ! -e "$work/missing-home/.local/bin/cerne"

mkdir -p "$work/bad/home/.local/bin/cerne"
if HOME=$work/bad/home TMPDIR=$tmp_root CERNE_INSTALL_ALLOW_FILE_URLS=1 CERNE_INSTALL_REPO_URL=file://$work/release CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=amd64 sh "$root/install.sh" >/dev/null 2>/dev/null; then
	echo "expected directory destination refusal" >&2
	exit 1
fi

mkdir -p "$work/link/home/.local/bin"
ln -s "$work/elsewhere" "$work/link/home/.local/bin/cerne"
if HOME=$work/link/home TMPDIR=$tmp_root CERNE_INSTALL_ALLOW_FILE_URLS=1 CERNE_INSTALL_REPO_URL=file://$work/release CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=amd64 sh "$root/install.sh" >/dev/null 2>/dev/null; then
	echo "expected symlink destination refusal" >&2
	exit 1
fi

mkdir -p "$work/link-parent/home" "$work/link-parent/outside"
ln -s "$work/link-parent/outside" "$work/link-parent/home/.local"
if HOME=$work/link-parent/home TMPDIR=$tmp_root CERNE_INSTALL_ALLOW_FILE_URLS=1 CERNE_INSTALL_REPO_URL=file://$work/release CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=amd64 sh "$root/install.sh" >/dev/null 2>/dev/null; then
	echo "expected symlink parent refusal" >&2
	exit 1
fi
test ! -e "$work/link-parent/outside/bin/cerne"

mkdir -p "$work/mismatch/download/v1.2.3"
cp "$work/release/download/v1.2.3/"*.tar.gz "$work/mismatch/download/v1.2.3/"
printf '%064d  cerne_v1.2.3_linux_amd64.tar.gz\n' 0 >"$work/mismatch/download/v1.2.3/checksums.txt"
if HOME=$work/mismatch-home TMPDIR=$tmp_root CERNE_INSTALL_ALLOW_FILE_URLS=1 CERNE_INSTALL_REPO_URL=file://$work/mismatch CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=amd64 sh "$root/install.sh" --version v1.2.3 >/dev/null 2>/dev/null; then
	echo "expected checksum mismatch failure" >&2
	exit 1
fi
test ! -e "$work/mismatch-home/.local/bin/cerne"

if HOME=$work/home TMPDIR=$tmp_root CERNE_INSTALL_ALLOW_FILE_URLS=1 CERNE_INSTALL_REPO_URL=file://$work/release CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=s390x sh "$root/install.sh" >/dev/null 2>/dev/null; then
	echo "expected unsupported architecture failure" >&2
	exit 1
fi

before=$(find "$tmp_root" -maxdepth 1 -name 'cerne-install.*' | wc -l | tr -d ' ')
test "$before" = "0"

for profile in .profile .zshrc .bashrc; do
	test ! -e "$work/home/$profile"
done

if grep -n 'sudo ' "$root/install.sh" >/dev/null; then
	echo "installer must not use sudo" >&2
	exit 1
fi
grep -- '--https-only' "$root/install.sh" >/dev/null

run_install --version v1.2.3 --agent codex >"$work/install-agent"
grep 'skill codex' "$work/install-agent" >/dev/null
run_install --version v1.2.3 --agent gemini >"$work/install-agent-gemini"
grep 'skill gemini' "$work/install-agent-gemini" >/dev/null

if run_install --version v1.2.4 --agent codex >"$work/install-skill-fail" 2>/dev/null; then
	echo "expected delegated skill failure" >&2
	exit 1
fi
grep 'cerne v1.2.4' "$work/install-skill-fail" >/dev/null

if run_install --agent unknown >/dev/null 2>/dev/null; then
	echo "expected unsupported agent failure" >&2
	exit 1
fi

if HOME=$work/no-file-override TMPDIR=$tmp_root CERNE_INSTALL_REPO_URL=file://$work/release CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=amd64 sh "$root/install.sh" >/dev/null 2>/dev/null; then
	echo "expected file URL refusal without explicit test override" >&2
	exit 1
fi

if HOME=$work/http-home TMPDIR=$tmp_root CERNE_INSTALL_REPO_URL=http://example.invalid/releases CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=amd64 sh "$root/install.sh" >/dev/null 2>/dev/null; then
	echo "expected non-HTTPS release URL refusal" >&2
	exit 1
fi

for bad_version in 1.2.3 v01.2.3 v1.02.3 v1.2.03 v1.2.3-01 v1.2.3-alpha. 'v1.2.3/evil'; do
	if run_install --version "$bad_version" >/dev/null 2>/dev/null; then
		echo "expected invalid version refusal: $bad_version" >&2
		exit 1
	fi
done

"$work/home/.local/bin/cerne" --version >"$work/before-invalid"
if run_install --version v1.2.5 >/dev/null 2>/dev/null; then
	echo "expected invalid downloaded binary failure" >&2
	exit 1
fi
"$work/home/.local/bin/cerne" --version >"$work/after-invalid"
cmp "$work/before-invalid" "$work/after-invalid"

if run_install --version v1.2.6 >/dev/null 2>/dev/null; then
	echo "expected archive symlink refusal" >&2
	exit 1
fi

if run_install --version v1.2.7 >/dev/null 2>/dev/null; then
	echo "expected archive directory refusal" >&2
	exit 1
fi

if run_install --version v1.2.8 >/dev/null 2>/dev/null; then
	echo "expected archive special-file refusal" >&2
	exit 1
fi

mkdir -p "$work/invalid-latest/latest/download"
cp "$work/release/download/v1.2.3/"*.tar.gz "$work/invalid-latest/latest/download/cerne_v01.2.3_linux_amd64.tar.gz"
sum=$(sha256sum "$work/invalid-latest/latest/download/cerne_v01.2.3_linux_amd64.tar.gz" | awk '{print $1}')
printf '%s  %s\n' "$sum" "cerne_v01.2.3_linux_amd64.tar.gz" >"$work/invalid-latest/latest/download/checksums.txt"
if HOME=$work/invalid-latest-home TMPDIR=$tmp_root CERNE_INSTALL_ALLOW_FILE_URLS=1 CERNE_INSTALL_REPO_URL=file://$work/invalid-latest CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=amd64 sh "$root/install.sh" >/dev/null 2>/dev/null; then
	echo "expected invalid latest archive name refusal" >&2
	exit 1
fi
