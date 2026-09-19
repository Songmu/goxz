#!/usr/bin/env bash
set -euo pipefail

goxz_version="v0.10.1"
goxz_bin="$(mktemp -d "${RUNNER_TEMP%/}/goxz-bin.XXXXXX")"
goxz_download="$(mktemp -d "${RUNNER_TEMP%/}/goxz-download.XXXXXX")"
trap 'rm -rf "$goxz_bin" "$goxz_download"' EXIT

case "${RUNNER_OS:-}" in
  Linux)
    target_os="linux"
    archive_format="tar.gz"
    ;;
  macOS)
    target_os="darwin"
    archive_format="zip"
    ;;
  Windows)
    target_os="windows"
    archive_format="zip"
    ;;
  *)
    echo "::error::unsupported runner operating system: ${RUNNER_OS:-unknown}"
    exit 1
    ;;
esac

case "${RUNNER_ARCH:-}" in
  X64)
    target_arch="amd64"
    ;;
  ARM64)
    target_arch="arm64"
    ;;
  *)
    echo "::error::unsupported runner architecture: ${RUNNER_ARCH:-unknown}"
    exit 1
    ;;
esac

archive_name="goxz_${goxz_version}_${target_os}_${target_arch}.${archive_format}"
archive="$goxz_download/$archive_name"
release_url="https://github.com/Songmu/goxz/releases/download/$goxz_version"
curl --fail --location --silent --show-error \
  --output "$archive" "$release_url/$archive_name"

if command -v gh >/dev/null && gh attestation verify --help >/dev/null 2>&1; then
  gh attestation verify "$archive" --repo Songmu/goxz
else
  checksums="$goxz_download/SHA256SUMS"
  curl --fail --location --silent --show-error \
    --output "$checksums" "$release_url/SHA256SUMS"
  expected_checksum="$(
    awk -v archive="$archive_name" '$2 == archive { print $1 }' "$checksums"
  )"
  if [[ -z "$expected_checksum" ]]; then
    echo "::error::checksum not found for $archive_name"
    exit 1
  fi
  if command -v sha256sum >/dev/null; then
    actual_checksum="$(sha256sum "$archive" | awk '{ print $1 }')"
  elif command -v shasum >/dev/null; then
    actual_checksum="$(shasum -a 256 "$archive" | awk '{ print $1 }')"
  else
    echo "::error::sha256sum or shasum is required to verify $archive_name"
    exit 1
  fi
  if [[ "$actual_checksum" != "$expected_checksum" ]]; then
    echo "::error::checksum mismatch for $archive_name"
    exit 1
  fi
fi

case "$archive_format" in
  tar.gz)
    tar -xzf "$archive" -C "$goxz_download"
    ;;
  zip)
    unzip -q "$archive" -d "$goxz_download"
    ;;
esac

executable="goxz"
if [[ "$target_os" == windows ]]; then
  executable="goxz.exe"
fi
install "$goxz_download/goxz_${goxz_version}_${target_os}_${target_arch}/$executable" \
  "$goxz_bin/$executable"

module_directory="$(cd "$GOXZ_DIRECTORY" && pwd -P)"
cd "$module_directory"
artifacts="$GOXZ_DESTINATION"
if [[ "$artifacts" != /* ]]; then
  artifacts="$module_directory/$artifacts"
fi

args=(-d "$artifacts")
if [[ -n "$GOXZ_PACKAGE_VERSION" ]]; then
  args+=(-pv "$GOXZ_PACKAGE_VERSION")
fi
if [[ -n "$GOXZ_OS" ]]; then
  args+=(-os "$GOXZ_OS")
fi
if [[ -n "$GOXZ_ARCH" ]]; then
  args+=(-arch "$GOXZ_ARCH")
fi
if [[ -n "$GOXZ_NAME" ]]; then
  args+=(-n "$GOXZ_NAME")
fi
if [[ -n "$GOXZ_INCLUDE" ]]; then
  args+=(-include "$GOXZ_INCLUDE")
fi
if [[ -n "$GOXZ_BUILD_LDFLAGS" ]]; then
  args+=(-build-ldflags "$GOXZ_BUILD_LDFLAGS")
fi
if [[ -n "$GOXZ_BUILD_TAGS" ]]; then
  args+=(-build-tags "$GOXZ_BUILD_TAGS")
fi

case "$GOXZ_STATIC" in
  true)
    args+=(-static)
    ;;
  false)
    ;;
  *)
    echo "::error::static must be either true or false"
    exit 1
    ;;
esac
case "$GOXZ_ZIP" in
  true)
    args+=(-z)
    ;;
  false)
    ;;
  *)
    echo "::error::zip must be either true or false"
    exit 1
    ;;
esac

read -r -a packages <<< "$GOXZ_PACKAGES"
"$goxz_bin/$executable" "${args[@]}" "${packages[@]}"

if [[ "${RUNNER_OS:-}" == Windows ]]; then
  artifacts="$(cygpath -m "$artifacts")"
fi
echo "artifacts=$artifacts" >> "$GITHUB_OUTPUT"
