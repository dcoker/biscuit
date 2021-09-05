#!/bin/bash
#
# Run a bunch of shell-based tests in isolated environments using Docker.
#
set -e

function aws() {
  docker run --network localstack -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test --rm amazon/aws-cli  --endpoint=http://localstack:4566 "$@"
}

export AWS_REGION=us-west-1
export REGION1=us-west-1
export REGION2=us-west-2

export AWS_ACCOUNT=$(aws --region=${REGION1} sts get-caller-identity | jq -r '.Account')
export KEY1=$(aws --region=${REGION1} kms create-key | jq -r '.KeyMetadata.KeyId')
export ARN1=arn:aws:kms:${REGION1}:${AWS_ACCOUNT}:key/${KEY1}
aws --region=${REGION1} kms create-alias --alias-name alias/biscuit-default --target-key-id ${ARN1}


export KEY2=$(aws --region=${REGION2} kms create-key | jq -r '.KeyMetadata.KeyId')
export ARN2=arn:aws:kms:${REGION2}:${AWS_ACCOUNT}:key/${KEY2}
aws --region=${REGION2} kms create-alias --alias-name alias/biscuit-default --target-key-id ${ARN2}

function invoke_one() {
  docker run \
    --network=localstack \
    -v $(pwd):/tests \
    -w /home/appuser \
    -e AWS_ACCESS_KEY_ID=test \
    -e AWS_SECRET_ACCESS_KEY=test \
    -e AWS_ENDPOINT=http://localstack:4566 \
    -e AWS_ACCOUNT=${AWS_ACCOUNT} \
    -e AWS_REGION=${AWS_REGION} \
    -e REGION1=${REGION1} \
    -e ARN1_REGION=${REGION1} \
    -e KEY1=${KEY1} \
    -e ARN1=${ARN1} \
    -e REGION2=${REGION2} \
    -e ARN2_REGION=${REGION2} \
    -e KEY2=${KEY2} \
    -e ARN2=${ARN2} \
    --entrypoint=/bin/bash \
    ghcr.io/dcoker/biscuit:latest \
    -c "$@"
}

RESULTS_DIR=$(mktemp -d)
echo ">>>>> Logging to ${RESULTS_DIR}"
for one_test in 0*sh; do
  echo ">>>>> Running test: ${one_test}"
  (invoke_one "/tests/${one_test}" && echo "+++++ PASSED" || echo "----- FAILED") \
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
