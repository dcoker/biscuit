#!/bin/bash -x
biscuit put -f store.yaml password god --key-id "${ARN1}","${ARN2}"
biscuit put -f store.yaml username oreilly
[[ "god" == "$(biscuit get -f store.yaml password)" ]]
[[ "oreilly" == "$(biscuit get -f store.yaml username)" ]]
