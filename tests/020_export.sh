#!/bin/bash -x
set -e
T=$(mktemp)
biscuit put -f store.yaml password god -a none
biscuit put -f store.yaml username oreilly --key-id "${ARN1}","${ARN2}" -a aesgcm256
biscuit put -f store.yaml spice scary --key-id "${ARN2}"
biscuit export -f store.yaml > "${T}"
[[ "3" == "$(wc -l "${T}" | awk '{print $1}')" ]]
grep ": god" "${T}"
grep ": oreilly" "${T}"
grep ": scary" "${T}"
