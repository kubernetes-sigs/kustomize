# Usage:  From repo root:
#  ./hacks/doGoMod.sh tidy
#  ./hacks/doGoMod.sh verify

operation=$1
for f in $(find ./ -name 'go.mod'); do
  echo $f
  d=$(dirname "$f")
  (cd $d; go mod $operation)
done
