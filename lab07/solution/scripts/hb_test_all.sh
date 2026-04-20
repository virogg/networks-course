#!/usr/bin/env bash
set -u
cd "$(dirname "$0")"

FAIL=0
for t in hb_test_death.sh hb_test_multi.sh hb_test_gap.sh; do
    echo "################ $t ################"
    if ! bash "./$t"; then
        FAIL=1
    fi
    echo
done

[[ $FAIL -eq 0 ]] && echo "ALL HB TESTS PASSED" || { echo "SOME HB TESTS FAILED"; exit 1; }
