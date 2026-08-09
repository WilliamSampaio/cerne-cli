#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
work=${TMPDIR:-/tmp}/cerne-install-test.$$
trap 'rm -rf "$work"' EXIT HUP INT TERM
mkdir -p "$work/release/download/v1.2.3" "$work/release/latest/download" "$work/src"

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

make_asset v1.2.3 linux amd64 "$work/release/download/v1.2.3"
cp "$work/release/download/v1.2.3/"* "$work/release/latest/download/"
mkdir -p "$work/release/download/v1.2.4"
make_asset v1.2.4 linux amd64 "$work/release/download/v1.2.4" 1

run_install() {
	HOME=$work/home PATH=/usr/bin:/bin CERNE_INSTALL_REPO_URL=file://$work/release CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=amd64 sh "$root/install.sh" "$@"
}

run_install >/tmp/cerne-install-out.$$
test -x "$work/home/.local/bin/cerne"
grep 'cerne v1.2.3' /tmp/cerne-install-out.$$ >/dev/null
grep 'add to PATH' /tmp/cerne-install-out.$$ >/dev/null

run_install >/dev/null

test ! -e "$work/home/.codex"
test ! -e "$work/home/.claude"

if HOME=$work/missing-home CERNE_INSTALL_REPO_URL=file://$work/release CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=amd64 sh "$root/install.sh" --version v9.9.9 >/dev/null 2>/dev/null; then
	echo "expected missing version failure" >&2
	exit 1
fi
test ! -e "$work/missing-home/.local/bin/cerne"

mkdir -p "$work/bad/home/.local/bin/cerne"
if HOME=$work/bad/home CERNE_INSTALL_REPO_URL=file://$work/release CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=amd64 sh "$root/install.sh" >/dev/null 2>/dev/null; then
	echo "expected directory destination refusal" >&2
	exit 1
fi

mkdir -p "$work/link/home/.local/bin"
ln -s "$work/elsewhere" "$work/link/home/.local/bin/cerne"
if HOME=$work/link/home CERNE_INSTALL_REPO_URL=file://$work/release CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=amd64 sh "$root/install.sh" >/dev/null 2>/dev/null; then
	echo "expected symlink destination refusal" >&2
	exit 1
fi

mkdir -p "$work/mismatch/download/v1.2.3"
cp "$work/release/download/v1.2.3/"*.tar.gz "$work/mismatch/download/v1.2.3/"
printf '%064d  cerne_v1.2.3_linux_amd64.tar.gz\n' 0 >"$work/mismatch/download/v1.2.3/checksums.txt"
if HOME=$work/mismatch-home CERNE_INSTALL_REPO_URL=file://$work/mismatch CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=amd64 sh "$root/install.sh" --version v1.2.3 >/dev/null 2>/dev/null; then
	echo "expected checksum mismatch failure" >&2
	exit 1
fi
test ! -e "$work/mismatch-home/.local/bin/cerne"

if HOME=$work/home CERNE_INSTALL_REPO_URL=file://$work/release CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=s390x sh "$root/install.sh" >/dev/null 2>/dev/null; then
	echo "expected unsupported architecture failure" >&2
	exit 1
fi

before=$(find "$work" -maxdepth 1 -name 'cerne-install.*' | wc -l | tr -d ' ')
test "$before" = "0"

for profile in .profile .zshrc .bashrc; do
	test ! -e "$work/home/$profile"
done

if grep -n 'sudo ' "$root/install.sh" >/dev/null; then
	echo "installer must not use sudo" >&2
	exit 1
fi

run_install --version v1.2.3 --agent codex >/tmp/cerne-install-agent.$$
grep 'skill codex' /tmp/cerne-install-agent.$$ >/dev/null
run_install --version v1.2.3 --agent gemini >/tmp/cerne-install-agent-gemini.$$
grep 'skill gemini' /tmp/cerne-install-agent-gemini.$$ >/dev/null

if run_install --version v1.2.4 --agent codex >/tmp/cerne-install-skill-fail.$$ 2>/dev/null; then
	echo "expected delegated skill failure" >&2
	exit 1
fi
grep 'cerne v1.2.4' /tmp/cerne-install-skill-fail.$$ >/dev/null

if run_install --agent unknown >/dev/null 2>/dev/null; then
	echo "expected unsupported agent failure" >&2
	exit 1
fi
