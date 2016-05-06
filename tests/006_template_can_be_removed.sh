#!/bin/bash -x
biscuit put -f store.yaml password god --key-id "${KEY1}"
biscuit put -f store.yaml username oreilly
sed -i s/_keys/_removed/g store.yaml
[[ "god" == "$(biscuit get -f store.yaml password)" ]]
[[ "oreilly" == "$(biscuit get -f store.yaml username)" ]]
