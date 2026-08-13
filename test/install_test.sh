#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
work=$(mktemp -d "${TMPDIR:-/tmp}/cerne-install-test.XXXXXX")
trap 'rm -rf "$work"' EXIT HUP INT TERM
tmp_root=$work/tmp
mkdir -p "$work/release/download/v1.2.3" "$work/release/latest/download" "$work/src" "$tmp_root"

fixture_sha256() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | awk '{print $1}'
	else
		shasum -a 256 "$1" | awk '{print $1}'
	fi
}

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
if test "\${1:-}" = "--platform"; then
  echo "$os/$arch"
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
	sum=$(fixture_sha256 "$dir/$archive")
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
		hardlink)
			printf 'not executable\n' >"$tmp/target"
			ln "$tmp/target" "$tmp/cerne"
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
	if test "$kind" = "hardlink"; then
		( cd "$tmp" && tar -czf "$dir/$archive" target cerne )
	else
		( cd "$tmp" && tar -czf "$dir/$archive" cerne )
	fi
	sum=$(fixture_sha256 "$dir/$archive")
	printf '%s  %s\n' "$sum" "$archive" >>"$dir/checksums.txt"
}

make_version_mismatch_asset() {
	version=$1
	reported=$2
	dir=$3
	tmp=$work/src/$version-linux-amd64-version-mismatch
	mkdir -p "$tmp"
	cat >"$tmp/cerne" <<EOF
#!/bin/sh
if test "\${1:-}" = "--version"; then
  echo "cerne $reported"
  exit 0
fi
exit 2
EOF
	chmod +x "$tmp/cerne"
	archive=cerne_${version}_linux_amd64.tar.gz
	( cd "$tmp" && tar -czf "$dir/$archive" cerne )
	sum=$(fixture_sha256 "$dir/$archive")
	printf '%s  %s\n' "$sum" "$archive" >>"$dir/checksums.txt"
}

for platform in linux-amd64 linux-arm64 darwin-amd64 darwin-arm64; do
	make_asset v1.2.3 "${platform%-*}" "${platform#*-}" "$work/release/download/v1.2.3"
done
cp "$work/release/download/v1.2.3/"* "$work/release/latest/download/"
mkdir -p "$work/release/download/v1.2.4"
make_asset v1.2.4 linux amd64 "$work/release/download/v1.2.4" 1
mkdir -p "$work/release/download/v1.2.3-rc.1"
make_asset v1.2.3-rc.1 linux amd64 "$work/release/download/v1.2.3-rc.1"
mkdir -p "$work/release/download/v1.2.5" "$work/release/download/v1.2.6" "$work/release/download/v1.2.7" "$work/release/download/v1.2.8" "$work/release/download/v1.2.9" "$work/release/download/v1.2.10"
make_bad_asset v1.2.5 invalid-binary "$work/release/download/v1.2.5"
make_bad_asset v1.2.6 symlink "$work/release/download/v1.2.6"
make_bad_asset v1.2.7 directory "$work/release/download/v1.2.7"
make_bad_asset v1.2.8 fifo "$work/release/download/v1.2.8"
make_version_mismatch_asset v1.2.9 v9.9.9 "$work/release/download/v1.2.9"
make_bad_asset v1.2.10 hardlink "$work/release/download/v1.2.10"

run_install() {
	HOME=$work/home PATH=/usr/bin:/bin TMPDIR=$tmp_root CERNE_INSTALL_TEST_MODE=1 CERNE_INSTALL_REPO_URL=file://$work/release CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=amd64 sh "$root/install.sh" "$@"
}

for platform in linux-amd64 linux-arm64 darwin-amd64 darwin-arm64; do
	matrix_home=$work/matrix/$platform/home
	matrix_os=${platform%-*}
	matrix_arch=${platform#*-}
	mkdir -p "$matrix_home"
	HOME=$matrix_home PATH=/usr/bin:/bin TMPDIR=$tmp_root CERNE_INSTALL_TEST_MODE=1 CERNE_INSTALL_REPO_URL=file://$work/release CERNE_INSTALL_OS=$matrix_os CERNE_INSTALL_ARCH=$matrix_arch sh "$root/install.sh" >/dev/null
	test "$("$matrix_home/.local/bin/cerne" --platform)" = "$matrix_os/$matrix_arch"
	HOME=$matrix_home PATH=/usr/bin:/bin TMPDIR=$tmp_root CERNE_INSTALL_TEST_MODE=1 CERNE_INSTALL_REPO_URL=file://$work/release CERNE_INSTALL_OS=$matrix_os CERNE_INSTALL_ARCH=$matrix_arch sh "$root/install.sh" >/dev/null
done

run_install >"$work/install-out"
test -x "$work/home/.local/bin/cerne"
grep 'cerne v1.2.3' "$work/install-out" >/dev/null
grep 'add to PATH' "$work/install-out" >/dev/null

run_install >/dev/null

test ! -e "$work/home/.codex"
test ! -e "$work/home/.claude"
test ! -e "$work/home/.gemini"

HOME=$work/prerelease-home PATH=/usr/bin:/bin TMPDIR=$tmp_root CERNE_INSTALL_TEST_MODE=1 CERNE_INSTALL_REPO_URL=file://$work/release CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=amd64 sh "$root/install.sh" >/dev/null
test "$("$work/prerelease-home/.local/bin/cerne" --version)" = "cerne v1.2.3"
HOME=$work/prerelease-home PATH=/usr/bin:/bin TMPDIR=$tmp_root CERNE_INSTALL_TEST_MODE=1 CERNE_INSTALL_REPO_URL=file://$work/release CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=amd64 sh "$root/install.sh" --version v1.2.3-rc.1 >/dev/null
test "$("$work/prerelease-home/.local/bin/cerne" --version)" = "cerne v1.2.3-rc.1"

if HOME=$work/missing-home TMPDIR=$tmp_root CERNE_INSTALL_TEST_MODE=1 CERNE_INSTALL_REPO_URL=file://$work/release CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=amd64 sh "$root/install.sh" --version v9.9.9 >/dev/null 2>/dev/null; then
	echo "expected missing version failure" >&2
	exit 1
fi
test ! -e "$work/missing-home/.local/bin/cerne"

mkdir -p "$work/bad/home/.local/bin/cerne"
if HOME=$work/bad/home TMPDIR=$tmp_root CERNE_INSTALL_TEST_MODE=1 CERNE_INSTALL_REPO_URL=file://$work/release CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=amd64 sh "$root/install.sh" >/dev/null 2>/dev/null; then
	echo "expected directory destination refusal" >&2
	exit 1
fi

mkdir -p "$work/link/home/.local/bin"
ln -s "$work/elsewhere" "$work/link/home/.local/bin/cerne"
if HOME=$work/link/home TMPDIR=$tmp_root CERNE_INSTALL_TEST_MODE=1 CERNE_INSTALL_REPO_URL=file://$work/release CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=amd64 sh "$root/install.sh" >/dev/null 2>/dev/null; then
	echo "expected symlink destination refusal" >&2
	exit 1
fi

mkdir -p "$work/link-parent/home" "$work/link-parent/outside"
ln -s "$work/link-parent/outside" "$work/link-parent/home/.local"
if HOME=$work/link-parent/home TMPDIR=$tmp_root CERNE_INSTALL_TEST_MODE=1 CERNE_INSTALL_REPO_URL=file://$work/release CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=amd64 sh "$root/install.sh" >/dev/null 2>/dev/null; then
	echo "expected symlink parent refusal" >&2
	exit 1
fi
test ! -e "$work/link-parent/outside/bin/cerne"

mkdir -p "$work/mismatch/download/v1.2.3"
cp "$work/release/download/v1.2.3/"*.tar.gz "$work/mismatch/download/v1.2.3/"
printf '%064d  cerne_v1.2.3_linux_amd64.tar.gz\n' 0 >"$work/mismatch/download/v1.2.3/checksums.txt"
if HOME=$work/mismatch-home TMPDIR=$tmp_root CERNE_INSTALL_TEST_MODE=1 CERNE_INSTALL_REPO_URL=file://$work/mismatch CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=amd64 sh "$root/install.sh" --version v1.2.3 >/dev/null 2>/dev/null; then
	echo "expected checksum mismatch failure" >&2
	exit 1
fi
test ! -e "$work/mismatch-home/.local/bin/cerne"

if HOME=$work/home TMPDIR=$tmp_root CERNE_INSTALL_TEST_MODE=1 CERNE_INSTALL_REPO_URL=file://$work/release CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=s390x sh "$root/install.sh" >/dev/null 2>/dev/null; then
	echo "expected unsupported architecture failure" >&2
	exit 1
fi

before=$(find "$tmp_root" -maxdepth 1 -name 'cerne-install.*' | wc -l | tr -d ' ')
test "$before" = "0"

mkdir -p "$work/hostile-target"
printf 'keep\n' >"$work/hostile-target/sentinel"
ln -s "$work/hostile-target" "$tmp_root/cerne-install.$$"
run_install >/dev/null
test -L "$tmp_root/cerne-install.$$"
test "$(cat "$work/hostile-target/sentinel")" = "keep"

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
run_install --version v1.2.3 --agent claude >"$work/install-agent-claude"
grep 'skill claude' "$work/install-agent-claude" >/dev/null
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

if HOME=$work/alternate-home TMPDIR=$tmp_root CERNE_INSTALL_REPO_URL=https://user:secret@example.invalid/releases sh "$root/install.sh" >/dev/null 2>"$work/alternate-error"; then
	echo "expected production release source override refusal" >&2
	exit 1
fi
if grep 'example.invalid\|user:secret' "$work/alternate-error" >/dev/null; then
	echo "release source refusal exposed the overridden URL" >&2
	exit 1
fi

mkdir -p "$work/downloaders"
cat >"$work/downloaders/curl" <<'EOF'
#!/bin/sh
set -eu
proto=0
tls=0
url=
out=
while test "$#" -gt 0; do
	case "$1" in
		--proto) test "$2" = "=https"; proto=1; shift 2 ;;
		--tlsv1.2) tls=1; shift ;;
		-o) out=$2; shift 2 ;;
		https://*) url=$1; shift ;;
		*) shift ;;
	esac
done
test "$proto$tls" = "11"
cp "$CERNE_TEST_RELEASE${url#https://fixtures.invalid}" "$out"
EOF
cat >"$work/downloaders/wget" <<'EOF'
#!/bin/sh
set -eu
if test "${1:-}" = "--help"; then
	echo --https-only
	exit 0
fi
https_only=0
url=
out=
while test "$#" -gt 0; do
	case "$1" in
		--https-only) https_only=1; shift ;;
		-O) out=$2; shift 2 ;;
		https://*) url=$1; shift ;;
		*) shift ;;
	esac
done
test "$https_only" = "1"
cp "$CERNE_TEST_RELEASE${url#https://fixtures.invalid}" "$out"
EOF
chmod +x "$work/downloaders/curl" "$work/downloaders/wget"
for downloader in curl wget; do
	HOME=$work/$downloader-home PATH=$work/downloaders:/usr/bin:/bin TMPDIR=$tmp_root CERNE_INSTALL_TEST_MODE=1 CERNE_INSTALL_REPO_URL=https://fixtures.invalid CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=amd64 CERNE_INSTALL_DOWNLOADER=$downloader CERNE_TEST_RELEASE=$work/release sh "$root/install.sh" >/dev/null
	test "$("$work/$downloader-home/.local/bin/cerne" --version)" = "cerne v1.2.3"
done

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

if run_install --version v1.2.10 >/dev/null 2>/dev/null; then
	echo "expected archive hardlink refusal" >&2
	exit 1
fi

if run_install --version v1.2.9 >/dev/null 2>/dev/null; then
	echo "expected binary version mismatch refusal" >&2
	exit 1
fi

mkdir -p "$work/missing-checksum/download/v1.2.3"
cp "$work/release/download/v1.2.3/cerne_v1.2.3_linux_amd64.tar.gz" "$work/missing-checksum/download/v1.2.3/"
printf '%064d  unrelated.tar.gz\n' 0 >"$work/missing-checksum/download/v1.2.3/checksums.txt"
"$work/home/.local/bin/cerne" --version >"$work/before-missing-checksum"
if HOME=$work/home PATH=/usr/bin:/bin TMPDIR=$tmp_root CERNE_INSTALL_TEST_MODE=1 CERNE_INSTALL_REPO_URL=file://$work/missing-checksum CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=amd64 sh "$root/install.sh" --version v1.2.3 >/dev/null 2>/dev/null; then
	echo "expected missing checksum entry failure" >&2
	exit 1
fi
"$work/home/.local/bin/cerne" --version >"$work/after-missing-checksum"
cmp "$work/before-missing-checksum" "$work/after-missing-checksum"

mkdir -p "$work/promotion-home/.local/bin" "$work/fail-promotion-bin"
cat >"$work/promotion-home/.local/bin/cerne" <<'EOF'
#!/bin/sh
test "${1:-}" = "--version" && echo "cerne old"
EOF
cat >"$work/fail-promotion-bin/mv" <<'EOF'
#!/bin/sh
exit 1
EOF
chmod +x "$work/promotion-home/.local/bin/cerne" "$work/fail-promotion-bin/mv"
if HOME=$work/promotion-home PATH=$work/fail-promotion-bin:/usr/bin:/bin TMPDIR=$tmp_root CERNE_INSTALL_TEST_MODE=1 CERNE_INSTALL_REPO_URL=file://$work/release CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=amd64 sh "$root/install.sh" --version v1.2.3 >/dev/null 2>/dev/null; then
	echo "expected promotion failure" >&2
	exit 1
fi
test "$("$work/promotion-home/.local/bin/cerne" --version)" = "cerne old"
test "$(find "$work/promotion-home/.local/bin" -maxdepth 1 -name '.cerne.*' | wc -l | tr -d ' ')" = "0"

mkdir -p "$work/invalid-latest/latest/download"
cp "$work/release/download/v1.2.3/cerne_v1.2.3_linux_amd64.tar.gz" "$work/invalid-latest/latest/download/cerne_v01.2.3_linux_amd64.tar.gz"
sum=$(fixture_sha256 "$work/invalid-latest/latest/download/cerne_v01.2.3_linux_amd64.tar.gz")
printf '%s  %s\n' "$sum" "cerne_v01.2.3_linux_amd64.tar.gz" >"$work/invalid-latest/latest/download/checksums.txt"
if HOME=$work/invalid-latest-home TMPDIR=$tmp_root CERNE_INSTALL_TEST_MODE=1 CERNE_INSTALL_REPO_URL=file://$work/invalid-latest CERNE_INSTALL_OS=linux CERNE_INSTALL_ARCH=amd64 sh "$root/install.sh" >/dev/null 2>/dev/null; then
	echo "expected invalid latest archive name refusal" >&2
	exit 1
fi

test "$(find "$tmp_root" -maxdepth 1 -type d -name 'cerne-install.*' | wc -l | tr -d ' ')" = "0"
