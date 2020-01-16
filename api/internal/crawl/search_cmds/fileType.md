Find all the documents having the `fileType` field set:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "exists": {
            "field": "fileType"
        }
    }
}
'
```

Find all the documents whose `fileType` field is not set:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
   "size": 10000,
    "query": {
        "bool": {
            "must_not": {
                "exists": {
                    "field": "fileType"
                }
            }
        }
    }
}
'
```

Search for all the documents whose `fileType` field is `resource`:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "filter": [
       { "regexp": { "fileType": "resource" }}
      ]
    }
  }
}
'
```

Count distinct values of the `fileType` field:
```
curl -X POST "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "aggs" : {
        "fileType_count" : {
            "cardinality" : {
                "field" : "fileType",
                "precision_threshold": 40000
            }
        }
    }
}
'
```

List all the values of the `fileType` field and the frequency of each value:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "aggs" : {
        "fileType" : {
            "terms" : {
              "field" : "fileType"
            }
        }
    }
}
'
```


For all the kustomization files in the index, list all the values of the
`fileType` field and the frequency of each value:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
             "filter": [
                { "regexp": { "filePath": ".*/kustomization((.yaml)?|(.yml)?)/*" }}
             ]
        }
    },
    "aggs" : {
        "fileType" : {
            "terms" : {
              "field" : "fileType"
            }
        }
    }
}
'
```

For all the non-kustomization files in the index, list all the values of the
`fileType` field and the frequency of each value:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
            "must_not": {
                "regexp": { "filePath": ".*/kustomization((.yaml)?|(.yml)?)/*" }
            }
        }
    },
    "aggs" : {
        "fileType" : {
            "terms" : {
              "field" : "fileType"
            }
        }
    }
}
'
```
