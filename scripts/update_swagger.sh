#!/bin/sh

# This script is used to cleanup swagger file
if [ "$#" -ne 1 ]; then
  echo "Usage: update_swagger.sh <filename>" >&2
  exit 1
fi

filename=$1
tmp_filename=$filename.tmp

# Extract version from CHANGELOG.md
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CHANGELOG_PATH="$PROJECT_ROOT/CHANGELOG.md"

# Get the latest version from CHANGELOG.md (first occurrence of ## [version])
if [ -f "$CHANGELOG_PATH" ]; then
  API_VERSION=$(grep -m 1 '^## \[' "$CHANGELOG_PATH" | sed 's/^## \[\([^]]*\)\].*/\1/')
  if [ -z "$API_VERSION" ]; then
    API_VERSION="version not set"
  fi
else
  API_VERSION="version not set"
fi

echo "Setting API version to: $API_VERSION"

# Update the API version and title in the info section
cat ${filename} | \
  jq --arg version "$API_VERSION" \
    '.info.version = $version | .info.title = "api-sdk"' > ${tmp_filename}
mv ${tmp_filename} ${filename}

# We also need to cleanup response code that are obsolete because
# either i) we override the error code or ii) we override success case where we respond with 201/204
cat ${filename} | \
  jq \
    'del(.paths[][]."responses"|select(has("400"))."default" ) | del(.paths[][]."responses"|select(has("201"))."200") | del(.paths[][]."responses"|select(has("204"))."200")' > ${tmp_filename}
mv ${tmp_filename} ${filename}

# Add an example value to the ConsistencyPreference to override the default of UNSPECIFIED being shown in the docs
cat ${filename} | \
  jq \
    '.definitions.ConsistencyPreference.example = "MINIMIZE_LATENCY"' > ${tmp_filename}
mv ${tmp_filename} ${filename}

# Finally, for 204, there should be no schema
cat ${filename} | \
  jq \
    'del(.paths[][]."responses"."204"."schema")' > ${tmp_filename}
mv ${tmp_filename} ${filename}
