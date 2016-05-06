#!/bin/bash -x
FAKE_KEY_ID=c06320d9-aaaa-aaaa-aaaa-08263b0789d5
biscuit put -f store.yaml password god --key-id ${FAKE_KEY_ID} 2>&1 | grep NotFoundException
