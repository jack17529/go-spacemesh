#!/usr/bin/env bash
set -e
# Ensure we use Go installed
errcho() {
    RED='\033[0;31m'
    NC='\033[0m' # no color
    echo -e "${RED}$1${NC}"
}

if ! (hash go 2>/dev/null) ; then
    errcho "Could not find a Go installation, please install Go and try again."
    exit 1;
fi

# Ensure we use Go version 1.14+
read major minor patch <<< $(go version | sed 's/go version go\([0-9]*\)\.\([0-9]*\)\.\([0-9]*\).*/\1 \2 \3/')
if [[ ${major} -ne 1 || ${minor} -lt 14 ]]; then
    errcho "Go 1.14+ is required (v$major.$minor.$patch is installed at `which go`)"
    exit 1;
fi
