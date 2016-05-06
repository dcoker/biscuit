#!/bin/bash -x
set -e
biscuit put -f store.yaml password god --key-id "${ARN1}","${ARN2}"
biscuit put -f store.yaml username oreilly

cp store.yaml corrupt1.yaml
sed -i "s@${ARN1_REGION}@xxx@g" corrupt1.yaml
[[ "god" == "$(biscuit get -f corrupt1.yaml password)" ]]
[[ "oreilly" == "$(biscuit get -f corrupt1.yaml username)" ]]

cp store.yaml corrupt2.yaml
sed -i "s@${ARN2_REGION}@xxx@g" corrupt2.yaml
[[ "god" == "$(biscuit get -f corrupt2.yaml password)" ]]
[[ "oreilly" == "$(biscuit get -f corrupt2.yaml username)" ]]
