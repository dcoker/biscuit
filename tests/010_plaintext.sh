#!/bin/bash -x
set -e
biscuit put -f store.yaml password god -a none
biscuit put -f store.yaml username oreilly

[[ "god" == "$(biscuit get -f store.yaml password)" ]]
[[ "oreilly" == "$(biscuit get -f store.yaml username)" ]]
