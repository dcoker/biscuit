#!/bin/bash -x
biscuit put -f corrupt1.yaml password god --key-id "${ARN1}","${ARN2}"
sed -i "s@${ARN1_REGION}@xxx@g" corrupt1.yaml
sed -i "s@${ARN2_REGION}@xxx@g" corrupt1.yaml
biscuit get -f corrupt1.yaml password
[[ 1 == "$?" ]]
