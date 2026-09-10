#!/usr/bin/env bash
set -eo pipefail

CONFIG=".proton/reviewbot/config.yaml"

GOTOOLCHAIN=auto go run golang.org/x/vuln/cmd/govulncheck@latest -json ./... > vulns.json

jq -r '.finding | select((.osv != null) and (.trace[0].function != null)) | .osv' < vulns.json > vulns_osv_ids.txt

# Read GO- prefixed ignore rules from reviewbot config (single source of truth)
if [ -f "$CONFIG" ]; then
    grep -oE 'GO-[0-9]{4}-[0-9]+' "$CONFIG" | while read -r id; do
        echo "ignoring $id (tracked in $CONFIG)"
        grep -v "$id" < vulns_osv_ids.txt > tmp || true
        mv tmp vulns_osv_ids.txt
    done
fi

# Fail if any unignored vulns remain
if [ -s vulns_osv_ids.txt ]; then
    while read -r osv; do
        jq --arg osvid "$osv" \
            '.osv | select(.id == $osvid) | {"id":.id, "ranges": .affected[0].ranges, "import": .affected[0].ecosystem_specific.imports[0].path}' \
            < vulns.json
    done < vulns_osv_ids.txt
    echo
    echo "Vulnerability found"
    exit 1
fi

echo
echo "No new vulnerabilities found."
