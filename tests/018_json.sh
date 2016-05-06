#!/bin/bash -x
set -e
[[ "1" == "$(biscuit list -f "${REPOSITORY}"/tests/single.json | wc -l)" ]]
[[ "bar" == "$(biscuit get -f "${REPOSITORY}"/tests/single.json name1)" ]]
