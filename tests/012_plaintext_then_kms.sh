#!/bin/bash -x
set -e
biscuit put -f store.yaml password god -a none
biscuit put -f store.yaml username oreilly --key-id "${ARN1}"
biscuit put -f store.yaml spice scary --key-id "${ARN1}"

[[ "god" == "$(biscuit get -f store.yaml password)" ]]
[[ "oreilly" == "$(biscuit get -f store.yaml username)" ]]
[[ "scary" == "$(biscuit get -f store.yaml spice)" ]]
cat store.yaml
