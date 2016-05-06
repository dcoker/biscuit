#!/bin/bash -x
biscuit put -f store.yaml password god
[[ 1 == "$?" ]]
