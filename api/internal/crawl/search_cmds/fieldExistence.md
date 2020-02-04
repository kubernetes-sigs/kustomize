Count the documents whose `document` field is empty (The reason why the `document` field
of a document is empty is because of empty documents):
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
   "size": 10000,
    "query": {
        "bool": {
            "must_not": {
                "exists": {
                    "field": "document"
                }
            }
        }
    }
}
'
```

Find all the documents having the `creationTime` field set:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "exists": {
            "field": "creationTime"
        }
    }
}
'
```

Find all the documents whose `creationTime` field is not set:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
   "size": 10000,
    "query": {
        "bool": {
            "must_not": {
                "exists": {
                    "field": "creationTime"
                }
            }
        }
    }
}
'
```

The following fields of a document in the kustomize index are always non-empty:
`repositoryUrl`, `filePath`, `defaultBranch`.

The following fields of a document in the kustomize index may be empty:
`kinds`, `identifiers`, `values`.
