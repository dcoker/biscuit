#!/bin/bash -x
set -e
biscuit put -f store.yaml password god --key-id "${ARN1}"
biscuit put -f store.yaml username oreilly
biscuit put -f store.yaml spice scary -a none

[[ "god" == "$(biscuit get -f store.yaml password)" ]]
[[ "oreilly" == "$(biscuit get -f store.yaml username)" ]]
[[ "scary" == "$(biscuit get -f store.yaml spice)" ]]
cat store.yaml
