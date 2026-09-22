#!/usr/bin/env bash
# End-to-end test against a moto server: seed a small AWS estate, import it
# with Unclick and check that `tofu plan` only imports.
#
# S3 is left out: the AWS provider reads bucket tags through S3 Control at
# <account-id>.<endpoint host>, which does not resolve for a local emulator.
#
# Needs: a moto server at $AWS_ENDPOINT_URL, python with boto3, tofu in PATH
# and an unclick binary built with the aws provider ($UNCLICK).
set -euo pipefail

here="$(cd "$(dirname "$0")" && pwd)"
UNCLICK="${UNCLICK:-unclick}"
export AWS_ENDPOINT_URL="${AWS_ENDPOINT_URL:-http://127.0.0.1:5000}"
export AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID:-test}"
export AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY:-test}"
export AWS_REGION="${AWS_REGION:-us-east-1}"
export AWS_DEFAULT_REGION="$AWS_REGION"

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

python "$here/seed.py"

cd "$work"
"$UNCLICK" import aws --regions "$AWS_REGION" --resources vpc,subnet,sg
cd generated/aws

tofu init -input=false -no-color >/dev/null
tofu plan -input=false -no-color | tee plan.txt
grep -E "^Plan: [1-9][0-9]* to import, 0 to add, 0 to change, 0 to destroy\.$" plan.txt
