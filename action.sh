#!/usr/bin/env bash
set -euo pipefail

goxz_version="v0.11.0"
goxz_bin="$(mktemp -d "${RUNNER_TEMP%/}/goxz-bin.XXXXXX")"
trap 'rm -rf "$goxz_bin"' EXIT

bash "$GITHUB_ACTION_PATH/install.sh" -b "$goxz_bin" "$goxz_version"

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
goxz="$goxz_bin/goxz"
if [[ "${RUNNER_OS:-}" == Windows ]]; then
  goxz+=".exe"
fi
"$goxz" "${args[@]}" "${packages[@]}"

if [[ "${RUNNER_OS:-}" == Windows ]]; then
  artifacts="$(cygpath -m "$artifacts")"
fi
echo "artifacts=$artifacts" >> "$GITHUB_OUTPUT"
