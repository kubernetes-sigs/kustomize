for f in $(find $1 -name '*.go'); do
  echo $f
  # go get golang.org/x/tools/cmd/goimports
  # {or} go run go.coder.com/go-tools/cmd/goimports
  goimports -w $f
done
