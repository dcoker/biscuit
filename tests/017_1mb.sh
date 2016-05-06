#!/bin/bash -x
set -e
dd if=/dev/urandom of=1mb.dat bs=1024 count=1000
EXPECTED=$(md5sum 1mb.dat | awk '{print $1}')
time biscuit put -f store.yaml 1mb-sb --from-file 1mb.dat --key-id "${ARN1}"
time biscuit put -f store.yaml 1mb-aes --from-file 1mb.dat -a aesgcm256
time biscuit put -f store.yaml 1mb-none --from-file 1mb.dat -a none
time biscuit get -f store.yaml 1mb-sb -o 1mb-sb-file.dat
time biscuit get -f store.yaml 1mb-aes -o 1mb-aes-file.dat
time biscuit get -f store.yaml 1mb-none -o 1mb-none-file.dat
time biscuit get -f store.yaml 1mb-sb > 1mb-sb-redir.dat
time biscuit get -f store.yaml 1mb-aes > 1mb-aes-redir.dat
time biscuit get -f store.yaml 1mb-none > 1mb-none-redir.dat
ls -l
[[ 7 == $(md5sum 1mb* | grep -c "${EXPECTED}") ]]
