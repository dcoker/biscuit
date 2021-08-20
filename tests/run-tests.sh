#!/bin/bash
#
# Run a bunch of shell-based tests in isolated environments using Docker.
#
set -e
REPOSITORY=/go/src/github.com/dcoker/biscuit/
if [ "${CONTINUOUS_INTEGRATION}" != "true" ]; then
  echo __ Running with credentials from biscuit-testing profile.
  AWS_ACCESS_KEY_ID="$(aws configure --profile biscuit-testing get aws_access_key_id)"
  AWS_SECRET_ACCESS_KEY="$(aws configure --profile biscuit-testing get aws_secret_access_key)"
fi
# Test keys created with `biscuit kms init`.
# user/cli-integration-test granted usage of keys via `biscuit kms edit-key-policy`.
AWS_ACCOUNT=872957446280
AWS_REGION=us-west-1
KEY1=8be1cc5b-3c70-4295-acad-e497dd421961
ARN1=arn:aws:kms:us-west-1:${AWS_ACCOUNT}:key/${KEY1}
ARN1_REGION=us-west-1
KEY2=37cca9a8-d2d0-43b8-9363-bb8320176cee
ARN2=arn:aws:kms:us-west-1:${AWS_ACCOUNT}:key/${KEY2}
ARN2_REGION=us-west-1

function invoke_one() {
  docker run -t \
    -e AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID}" \
    -e AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY}" \
    -e AWS_REGION="${AWS_REGION}" \
    -e KEY1="${KEY1}" \
    -e ARN1="${ARN1}" \
    -e ARN1_REGION="${ARN1_REGION}" \
    -e KEY2="${KEY2}" \
    -e ARN2="${ARN2}" \
    -e ARN2_REGION="${ARN2_REGION}" \
    -e REPOSITORY=${REPOSITORY} \
    -w /tmp \
    biscuit/local \
    /bin/bash -c "$@"
}

RESULTS_DIR=$(mktemp -d)
echo ">>>>> Logging to ${RESULTS_DIR}"
for one_test in 0*sh; do
  echo ">>>>> Running test: ${one_test}"
  (invoke_one "${REPOSITORY}/tests/${one_test}" && echo "+++++ PASSED" || echo "----- FAILED") \
    > "${RESULTS_DIR}/${one_test}.log" \
    2>&1 &
done
echo ">> waiting"
wait
for test_log in "${RESULTS_DIR}"/*.log; do
  echo -n "${test_log}: "
  if grep -q '+++++ PASSED' "${test_log}"; then
   echo PASSED
  else
    echo "FAILED :("
    cat "${test_log}"
  fi
done
