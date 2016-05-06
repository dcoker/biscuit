#!/bin/bash -x
set -e
biscuit get -f 404.yaml key 2>&1 | grep 'no such file'
