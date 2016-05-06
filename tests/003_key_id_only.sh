#!/bin/bash -x
biscuit put -f store.yaml password god --key-id "${KEY1}"
[[ "god" == "$(biscuit get -f store.yaml password)" ]]
