#!/bin/bash -x
biscuit put -f store.yaml password god --key-id "${ARN1}"
[[ "god" == "$(biscuit get -f store.yaml password)" ]]
