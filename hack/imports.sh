for f in $(find ./ -name '*.go'); do
  echo $f
  # go run go.coder.com/go-tools/cmd/goimports
  ~/gopath/bin/goimports -w $f
done
