for f in $(find ./ -name '*.go'); do
  echo $f
  ~/gopath/bin/goimports -w $f
done
