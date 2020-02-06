Find all the documents having the `user` field set:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
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
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
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
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
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
curl -s -X POST "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
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
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "aggs" : {
        "user" : {
            "terms" : {
                "field" : "user",
                "size" : 20
            }
        }
    }
}
'
```

Count distinct values of the `user` field for all the kustomization files in the index:
```
curl -s -X POST "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
             "filter": [
                { "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }}
             ]
        }
    },
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

For all the kustomization files in the index, list all the values of the
`user` field and the frequency of each value:
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
        "user" : {
            "terms" : {
                "field" : "user",
                "size": 20
            }
        }
    }
}
'
```

Count distinct values of the `user` field for all the kustomize resource files in the index:
```
curl -s -X POST "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
            "must_not": {
                "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }
            },
            "filter": [
                { "regexp": { "fileType": "resource" }}
            ]
        }
    },
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

For all the kustomize resource files in the index, list all the values of the
`user` field and the frequency of each value:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
            "must_not": {
                "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }
            },
            "filter": [
                { "regexp": { "fileType": "resource" }}
            ]
        }
    },
    "aggs" : {
        "user" : {
            "terms" : {
                "field" : "user",
                "size": 20
            }
        }
    }
}
'
```

Count distinct values of the `user` field for all the kustomize generator files in the index:
```
curl -s -X POST "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
            "must_not": {
                "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }
            },
            "filter": [
                { "regexp": { "fileType": "generator" }}
            ]
        }
    },
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

For all the kustomize generator files in the index, list all the values of the
`user` field and the frequency of each value:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
            "must_not": {
                "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }
            },
            "filter": [
                { "regexp": { "fileType": "generator" }}
            ]
        }
    },
    "aggs" : {
        "user" : {
            "terms" : {
              "field" : "user",
              "size": 20  
            }
        }
    }
}
'
```

Count distinct values of the `user` field for all the kustomize transformer files in the index:
```
curl -s -X POST "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
            "must_not": {
                "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }
            },
            "filter": [
                { "regexp": { "fileType": "transformer" }}
            ]
        }
    },
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

For all the kustomize transformer files in the index, list all the values of the
`user` field and the frequency of each value:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
            "must_not": {
                "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }
            },
            "filter": [
                { "regexp": { "fileType": "transformer" }}
            ]
        }
    },
    "aggs" : {
        "user" : {
            "terms" : {
                "field" : "user",
                "size": 20
            }
        }
    }
}
'
```

Count distinct values of the `user` field for all the kustomize generator dirs in the index:
```
curl -s -X POST "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
            "filter": [
                { "regexp": { "fileType": "generator" }},
                { "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }}
            ]
        }
    },
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

For all the kustomize generator dirs in the index, list all the values of the
`user` field and the frequency of each value:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
            "filter": [
                { "regexp": { "fileType": "generator" }},
                { "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }}
            ]
        }
    },
    "aggs" : {
        "user" : {
            "terms" : {
              "field" : "user",
              "size": 20  
            }
        }
    }
}
'
```

Count distinct values of the `user` field for all the kustomize transformer dirs in the index:
```
curl -s -X POST "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
            "filter": [
                { "regexp": { "fileType": "transformer" }},
                { "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }}
            ]
        }
    },
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

For all the kustomize transformer dirs in the index, list all the values of the
`user` field and the frequency of each value:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
            "filter": [
                { "regexp": { "fileType": "transformer" }},
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