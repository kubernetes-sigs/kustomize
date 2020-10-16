if ! command -v jq &> /dev/null ; then
    echo Please install jq
    echo on ubuntu: sudo apt-get install jq
    exit 1
fi

GOPATH=$(go env GOPATH)

OPENAPIINFO=$(jq -r '.info' openapi/kubernetesapi/swagger.json | sed 's/[\" *]//g' | tr -d '\n')
sed -i "s/Info = \".*\"/Info = \"$OPENAPIINFO\"/g" 'openapi/kubernetesapi/openapiinfo.go'

(go get -u github.com/go-bindata/go-bindata/...)
$GOPATH/bin/go-bindata --pkg kubernetesapi -o openapi/kubernetesapi/swagger.go openapi/kubernetesapi/swagger.json
$GOPATH/bin/go-bindata --pkg kustomizationapi -o openapi/kustomizationapi/swagger.go openapi/kustomizationapi/swagger.json
