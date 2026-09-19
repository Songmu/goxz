#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<EOF
Usage: $0 [-b bindir] version

Install a goxz release into bindir. The default bindir is ./bin.
EOF
  exit 2
}

bindir="./bin"
while getopts "b:h" option; do
  case "$option" in
    b)
      bindir="$OPTARG"
      ;;
    h)
      usage
      ;;
    *)
      usage
      ;;
  esac
done
shift $((OPTIND - 1))

if [[ "$#" -ne 1 ]]; then
  usage
fi
version="$1"

case "${RUNNER_OS:-$(uname -s)}" in
  Linux)
    target_os="linux"
    archive_format="tar.gz"
    ;;
  macOS | Darwin)
    target_os="darwin"
    archive_format="zip"
    ;;
  Windows | MINGW* | MSYS* | CYGWIN*)
    target_os="windows"
    archive_format="zip"
    ;;
  *)
    echo "unsupported operating system: ${RUNNER_OS:-$(uname -s)}" >&2
    exit 1
    ;;
esac

case "${RUNNER_ARCH:-$(uname -m)}" in
  X64 | x86_64)
    target_arch="amd64"
    ;;
  ARM64 | arm64 | aarch64)
    target_arch="arm64"
    ;;
  *)
    echo "unsupported architecture: ${RUNNER_ARCH:-$(uname -m)}" >&2
    exit 1
    ;;
esac

tmp_base="${RUNNER_TEMP:-${TMPDIR:-/tmp}}"
download_dir="$(mktemp -d "${tmp_base%/}/goxz-download.XXXXXX")"
trap 'rm -rf "$download_dir"' EXIT

archive_name="goxz_${version}_${target_os}_${target_arch}.${archive_format}"
archive="$download_dir/$archive_name"
release_url="https://github.com/Songmu/goxz/releases/download/$version"
curl --fail --location --silent --show-error \
  --output "$archive" "$release_url/$archive_name"

gh_supports_safe_attestation() {
  local gh_version
  gh_version="$(gh --version | awk 'NR == 1 { print $3 }')"
  if [[ ! "$gh_version" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+) ]]; then
    return 1
  fi
  if ((BASH_REMATCH[1] > 2)); then
    return 0
  fi
  if ((BASH_REMATCH[1] < 2)); then
    return 1
  fi
  ((BASH_REMATCH[2] >= 93))
}

if command -v gh >/dev/null &&
  gh_supports_safe_attestation &&
  gh attestation verify --help >/dev/null 2>&1; then
  gh attestation verify "$archive" --repo Songmu/goxz
else
  checksums="$download_dir/SHA256SUMS"
  curl --fail --location --silent --show-error \
    --output "$checksums" "$release_url/SHA256SUMS"
  expected_checksum="$(
    awk -v archive="$archive_name" '$2 == archive { print $1 }' "$checksums"
  )"
  if [[ -z "$expected_checksum" ]]; then
    echo "checksum not found for $archive_name" >&2
    exit 1
  fi
  if command -v sha256sum >/dev/null; then
    actual_checksum="$(sha256sum "$archive" | awk '{ print $1 }')"
  elif command -v shasum >/dev/null; then
    actual_checksum="$(shasum -a 256 "$archive" | awk '{ print $1 }')"
  else
    echo "sha256sum or shasum is required to verify $archive_name" >&2
    exit 1
  fi
  if [[ "$actual_checksum" != "$expected_checksum" ]]; then
    echo "checksum mismatch for $archive_name" >&2
    exit 1
  fi
fi

case "$archive_format" in
  tar.gz)
    tar -xzf "$archive" -C "$download_dir"
    ;;
  zip)
    unzip -q "$archive" -d "$download_dir"
    ;;
esac

executable="goxz"
if [[ "$target_os" == windows ]]; then
  executable="goxz.exe"
fi
install -d "$bindir"
install "$download_dir/goxz_${version}_${target_os}_${target_arch}/$executable" \
  "$bindir/$executable"
