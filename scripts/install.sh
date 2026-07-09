#!/usr/bin/env sh
set -eu

release_dir="${1:?usage: scripts/install.sh <release-dir> <install-dir> [artifact]}"
install_dir="${2:?usage: scripts/install.sh <release-dir> <install-dir> [artifact]}"
artifact="${3:-recomphamr_linux_amd64.tar.gz}"
manifest="$release_dir/SHA256SUMS"
artifact_path="$release_dir/$artifact"

if [ ! -f "$artifact_path" ]; then
  echo "artifact not found: $artifact_path" >&2
  exit 1
fi

if [ ! -f "$manifest" ]; then
  echo "checksum manifest not found: $manifest" >&2
  exit 1
fi

expected="$(awk -v name="$artifact" '$2 == name || $2 == "*" name { print tolower($1); exit }' "$manifest")"
if [ -z "$expected" ]; then
  echo "artifact is missing from SHA256SUMS: $artifact" >&2
  exit 1
fi

actual="$(sha256sum "$artifact_path" | awk '{ print tolower($1) }')"
if [ "$actual" != "$expected" ]; then
  echo "checksum mismatch for $artifact" >&2
  exit 1
fi

mkdir -p "$install_dir"
tar -xzf "$artifact_path" -C "$install_dir"

if [ ! -f "$install_dir/recomphamr" ]; then
  echo "installed archive did not contain recomphamr" >&2
  exit 1
fi

chmod +x "$install_dir/recomphamr"
echo "Installed RecompHamr to $install_dir/recomphamr"
