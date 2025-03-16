#!/usr/bin/env bash

set -e

FILE="$GOFILE"
HEADER_FILE=$(mktemp)

cat <<EOF > "$HEADER_FILE"
// @Header       all {string} X-Request-ID "UUID of the request"
// @Header       all {string} X-API-Version "API version, e.g. v1alpha"
// @Header       all {int} X-Ratelimit-Limit "Rate limit value"
// @Header       all {int} X-Ratelimit-Remaining "Rate limit remaining"
// @Header       all {int} X-Ratelimit-Reset "Rate limit reset interval in seconds"
EOF

if [[ -z "$FILE" || ! -f "$FILE" ]]; then
    echo "Error: File $FILE does not exist or is not set!"
    exit 1
fi

if [[ "$(uname)" == "Darwin" ]]; then
    # On macOS, prefer gsed
    if ! command -v gsed >/dev/null 2>&1; then
        echo "Error: gsed not found. Please install it with 'brew install gnu-sed'"
        exit 1
    fi
    SED="gsed"
else
    SED="sed"
fi

"$SED" -i "/\/\/ @COMMON_HEADERS_PLACEHOLDER/r $HEADER_FILE" "$FILE"
"$SED" -i "/\/\/ @COMMON_HEADERS_PLACEHOLDER/d" "$FILE"
