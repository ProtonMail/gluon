#!/usr/bin/env bash

set -eo pipefail

main(){
    local go_package="$1"
    govulncheck -json "$go_package" > vulns.json

    jq -r '.finding | select( (.osv != null) and (.trace[0].function != null) ) | .osv ' < vulns.json > vulns_osv_ids.txt

    ignore GO-2026-4866 "BRIDGE-525 crypto/x509 verifying a certificate chain excluded DNS constraints which were not applied to wildcard DNS SANs."
    ignore GO-2026-4870 "BRIDGE-525 crypto/tls if one side of TLS connection sends multiple key messages post handshake can lead to deadlock."
    ignore GO-2026-4946 "BRIDGE-525 crypto/x509 validating certificate chains is unexpectedly inefficient when chains contain very large number of policy mappings."
    ignore GO-2026-4947 "BRIDGE-525 crypto/x509 during chain building the amount of work is not limited passed in VerifyOptions.Intermediates which can lead to denial of service." 

    has_vulns

    echo
    echo "No new vulnerabilities found."
}

ignore(){
    echo "ignoring $1 fix: $2"
    cp vulns_osv_ids.txt tmp
    grep -v "$1" < tmp > vulns_osv_ids.txt || true
    rm tmp
}

has_vulns(){
    has=false
    while read -r osv; do
        jq \
            --arg osvid "$osv" \
            '.osv | select ( .id == $osvid) | {"id":.id, "ranges": .affected[0].ranges,  "import": .affected[0].ecosystem_specific.imports[0].path}' \
            < vulns.json
        has=true
    done < vulns_osv_ids.txt

    if [ "$has" == true ]; then
        echo
        echo "Vulnerability found"
        return 1
    fi
}

main
