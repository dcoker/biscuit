#!/bin/bash -x
set -e
biscuit put -f store.yaml password god -a none
biscuit put -f store.yaml username oreilly --key-id "${ARN1}","${ARN2}" -a aesgcm256
biscuit put -f store.yaml spice scary --key-id "${ARN2}"

[[ "god" == "$(biscuit get -f store.yaml password)" ]]
[[ "oreilly" == "$(biscuit get -f store.yaml username)" ]]
[[ "scary" == "$(biscuit get -f store.yaml spice)" ]]
grep aesgcm256 store.yaml
grep secretbox store.yaml
grep none store.yaml
