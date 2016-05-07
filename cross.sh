#!/bin/bash
set -e
for rls in build/biscuit/{linux,darwin}*; do \
  tar czf build/biscuit-$(echo ${rls} | cut -f3 -d/).tgz -C ${rls} biscuit; \
done
for rls in build/biscuit/windows*; do \
  mv ${rls}/biscuit.exe build/biscuit-$(echo ${rls} | cut -f3 -d/).exe
done
ls -l build/
