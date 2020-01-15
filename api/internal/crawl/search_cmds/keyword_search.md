Count the documents in the index whose `repositoryUrl` field starts with
`https://github.com/`:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "filter": [
       { "regexp": { "repositoryUrl": "https://github.com/.*" }}
      ]
    }
  }
}
'
```

Count the documents in the index whose `repositoryUrl` field does not start with
`https://github.com/`:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "must_not": [
       { "regexp": { "repositoryUrl": "https://github.com/.*" }}
      ]
    }
  }
}
'
```

Search all the documents matching the given `repositoryUrl` and `filePath`, and return 
a version for each search hit:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "size": 10000,
  "version": true,
  "query": {
    "bool": {
      "filter": [
       { "regexp": { "repositoryUrl": "git@github.com:talos-systems/talos-controller-manager" }},
       { "regexp": { "filePath": "hack/config.*" }}
      ]
    }
  }
}
'
```

Search all the documents whose filePath ends with one of these following three filenames:
`kustomization.yaml`, `kustomization.yml`, `kustomization`:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "filter": [
       { "regexp": { "filePath": ".*/kustomization((.yaml)?|(.yml)?)/*" }}
      ]
    }
  }
}
'
```

Search all the documents whose filePath does not end with any of these following
three filenames: `kustomization.yaml`, `kustomization.yml`, `kustomization`:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "must_not": [
       { "regexp": { "filePath": ".*/kustomization((.yaml)?|(.yml)?)/*" }}
      ]
    }
  }
}
'
```