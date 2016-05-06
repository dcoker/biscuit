#!/bin/bash
#
# Run a bunch of shell-based tests in isolated environments using Docker.
#
set -e
REPOSITORY=/go/src/github.com/dcoker/secrets/
AWS_ACCESS_KEY_ID="$(aws configure get aws_access_key_id)"
AWS_SECRET_ACCESS_KEY="$(aws configure get aws_secret_access_key)"
AWS_SESSION_TOKEN="$(aws configure get aws_session_token)"
AWS_REGION=us-west-1
KEY1=37793df5-ad32-4d06-b19f-bfb95cee4a35
ARN1=arn:aws:kms:us-west-1:105770556716:key/${KEY1}
ARN1_REGION=us-west-1
KEY2=c0045b15-9880-4b17-84da-a35760e8a16f
ARN2=arn:aws:kms:us-west-2:105770556716:key/${KEY2}
ARN2_REGION=us-west-2

function invoke_one() {
  docker run -t \
    -e AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID}" \
    -e AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY}" \
    -e AWS_SESSION_TOKEN="${AWS_SESSION_TOKEN}" \
    -e AWS_REGION="${AWS_REGION}" \
    -e KEY1="${KEY1}" \
    -e ARN1="${ARN1}" \
    -e ARN1_REGION="${ARN1_REGION}" \
    -e KEY2="${KEY2}" \
    -e ARN2="${ARN2}" \
    -e ARN2_REGION="${ARN2_REGION}" \
    -e REPOSITORY=${REPOSITORY} \
    -w /tmp \
    secrets/local \
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
for test_log in ${RESULTS_DIR}/*.log; do
  echo -n "${test_log}: "
  if grep -q '+++++ PASSED' "${test_log}"; then
   echo PASSED
  else
    echo "FAILED :("
    cat "${test_log}"
  fi
done
