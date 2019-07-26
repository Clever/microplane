set -e

t="Verify Microplane commands work e2e on GitLab"

echo "== START: ($t) =="

echo ""
echo "init"
echo ""

./bin/mp init "mp-test-1"

echo ""
echo "clone"
echo ""

./bin/mp clone

echo ""
echo "plan"
echo ""

./bin/mp plan -b test1 -m "test1" -- touch test-$(date +"%T").txt

echo ""
echo "push"
echo ""

./bin/mp push -a microplane-gitlab

echo ""
echo "merge"
echo ""

./bin/mp merge --ignore-build-status

echo ""
echo "== END: ($t) =="
