#!/bin/bash -x
set -e
AWS_REGION="${ARN2_REGION}" biscuit put -f store.yaml password r1 --key-id "${ARN1}"
AWS_REGION="${ARN1_REGION}" biscuit put -f store.yaml username r2 --key-id "${ARN2}"
[[ "r1" == "$(AWS_REGION="${ARN2_REGION}" biscuit get -f store.yaml password)" ]]
[[ "r2" == "$(AWS_REGION="${ARN1_REGION}" biscuit get -f store.yaml username)" ]]
cat store.yaml
