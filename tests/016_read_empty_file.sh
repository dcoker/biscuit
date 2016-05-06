#!/bin/bash -x
touch empty.yaml
biscuit get -f empty.yaml key
[[ 1 == "$?" ]]
