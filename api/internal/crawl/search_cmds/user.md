Find all the documents having the `user` field set:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "exists": {
            "field": "user"
        }
    }
}
'
```

Find all the documents whose `user` field is not set:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
   "size": 10000,
    "query": {
        "bool": {
            "must_not": {
                "exists": {
                    "field": "user"
                }
            }
        }
    }
}
'
```

Search for all the documents whose `user` field is `kubernetes-sigs`:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "filter": [
       { "regexp": { "user": "kubernetes-sigs" }}
      ]
    }
  }
}
'
```

Count distinct values of the `user` field:
```
curl -X POST "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "aggs" : {
        "user_count" : {
            "cardinality" : {
                "field" : "user",
                "precision_threshold": 40000
            }
        }
    }
}
'
```

List all the values of the `user` field and the frequency of each value:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "aggs" : {
        "user" : {
            "terms" : {
              "field" : "user"
            }
        }
    }
}
'
```


For all the kustomization files in the index, list all the values of the
`user` field and the frequency of each value:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
             "filter": [
                { "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }}
             ]
        }
    },
    "aggs" : {
        "user" : {
            "terms" : {
              "field" : "user"
            }
        }
    }
}
'
```

For all the non-kustomization files in the index, list all the values of the
`user` field and the frequency of each value:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
            "must_not": {
                "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }
            }
        }
    },
    "aggs" : {
        "user" : {
            "terms" : {
              "field" : "user"
            }
        }
    }
}
'
```
