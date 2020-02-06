Find all the documents having the `fileType` field set:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
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
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
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
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
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

Search for all the kustomization files whose `fileType` field is `resource`:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "filter": [
       { "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }}, 
       { "regexp": { "fileType": "resource" }}
      ]
    }
  }
}
'
```

Search for all the kustomize resource files:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "filter": [
       { "regexp": { "fileType": "resource" }}
      ],
      "must_not": {
        "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }
      }
    }
  }
}
'
```

Search all the kustomization files including a `generators` field:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "size": 10000,
  "query": {
    "bool": {
      "must": {
        "match" : {
          "identifiers" : {
            "query" : "generators"
          }
        }
      },
      "filter": {
        "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }
      }
    }
  }
}
'
```

Search for all the documents whose `fileType` field is `generator`:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "filter": [
       { "regexp": { "fileType": "generator" }}
      ]
    }
  }
}
'
```

Search for all the kustomization files whose `fileType` field is `generator`:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "filter": [
       { "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }}, 
       { "regexp": { "fileType": "generator" }}
      ]
    }
  }
}
'
```

Search for all the kustomize generator files:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "filter": [
       { "regexp": { "fileType": "generator" }}
      ],
      "must_not": {
        "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }
      }
    }
  }
}
'
```

Search all the kustomization files including a `transformers` field:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "size": 10000,
  "query": {
    "bool": {
      "must": {
        "match" : {
          "identifiers" : {
            "query" : "transformers"
          }
        }
      },
      "filter": {
        "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }
      }
    }
  }
}
'
```

Search for all the documents whose `fileType` field is `transformer`:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "filter": [
       { "regexp": { "fileType": "transformer" }}
      ]
    }
  }
}
'
```

Search for all the kustomization files whose `fileType` field is `transformer`:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "filter": [
       { "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }}, 
       { "regexp": { "fileType": "transformer" }}
      ]
    }
  }
}
'
```

Search for all the kustomize transformer files:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "filter": [
       { "regexp": { "fileType": "transformer" }}
      ],
      "must_not": {
        "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }
      }
    }
  }
}
'
```

Count distinct values of the `fileType` field:
```
curl -s -X POST "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
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
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
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
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
             "filter": [
                { "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }}
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
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
            "must_not": {
                "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }
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
