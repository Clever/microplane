# !/bin/bash
#
# ./integration_test.sh
#
# This is a prototype end-to-end test for microplane.
#
# Please read carefully and use at your own risk.

set -e

####################
# Script setup
####################
provider=$1
nuke=$2

usage() {
  echo "usage: ./integration_test.sh <github|gitlab> {nuke}"
}

# specify provider
if [ "$provider" != "github" ] && [ "$provider" != "gitlab" ]; then
  usage
  exit 1
fi

owner="microplane-test"
if [ "$provider" == "gitlab" ]; then
  owner="microplane-gitlab"
fi

# nuke is optional
if [ "$nuke" == "nuke" ]; then
  echo "nuking ./mp"
  rm -rf ./mp
fi

if [ -d "mp" ]; then
    echo "Working directory ./mp already exists. Please remove before running."
    exit 1
fi

####################
# Run microplane e2e
####################

echo "[Init]"
tmpfile=$(mktemp /tmp/mp-integration-test.XXXXXX)
echo "$owner/1" >> $tmpfile
echo "$owner/2" >> $tmpfile

./bin/mp init --provider $provider -f $tmpfile
rm $tmpfile

echo "[Clone]"
./bin/mp clone
ts=`date +"%T"`


echo "[Plan]"
./bin/mp plan -b plan -m "plan" -- sh -c "echo $ts >> README.md"

echo "[Push]"
./bin/mp push --throttle 2s -a nathanleiby

echo "[Merge]"
cmd='./bin/mp merge --throttle 2s --ignore-build-status --ignore-review-approval'
duration=10
until $cmd; do
    echo "waiting a bit ($duration seconds) so PRs are mergeable..."
    sleep $duration
done

