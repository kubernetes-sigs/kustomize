Count distinct values of the `repositoryUrl` field:
```
curl -s -X POST "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "aggs" : {
        "repositoryUrl_count" : {
            "cardinality" : {
                "field" : "repositoryUrl",
                "precision_threshold": 40000
            }
        }
    }
}
'
```

Count how many Github repositories include kustomization files:
```
curl -s -X POST "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
             "filter": [
                { "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*" }}
             ]
        }
    },
    "aggs" : {
        "repositoryUrl_count" : {
            "cardinality" : {
                "field" : "repositoryUrl",
                "precision_threshold": 40000
            }
        }
    }
}
'
```

Count distinct values of the `repositoryUrl` field for all the kustomize resource files in the index:
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
        "repositoryUrl_count" : {
            "cardinality" : {
                "field" : "repositoryUrl",
                "precision_threshold": 40000
            }
        }
    }
}
'
```

Count distinct values of the `repositoryUrl` field for all the kustomize generator files in the index:
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
        "repositoryUrl_count" : {
            "cardinality" : {
                "field" : "repositoryUrl",
                "precision_threshold": 40000
            }
        }
    }
}
'
```

Count distinct values of the `repositoryUrl` field for all the kustomize transformer files in the index:
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
        "repositoryUrl_count" : {
            "cardinality" : {
                "field" : "repositoryUrl",
                "precision_threshold": 40000
            }
        }
    }
}
'
```

Count distinct values of the `repositoryUrl` field for all the kustomize resource dirs in the index:
```
curl -s -X POST "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
            "filter": [
                { "regexp": { "fileType": "resource" }},
                { "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }}
            ]
        }
    },
    "aggs" : {
        "repositoryUrl_count" : {
            "cardinality" : {
                "field" : "repositoryUrl",
                "precision_threshold": 40000
            }
        }
    }
}
'
```

Count distinct values of the `repositoryUrl` field for all the kustomize generator dirs in the index:
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
        "repositoryUrl_count" : {
            "cardinality" : {
                "field" : "repositoryUrl",
                "precision_threshold": 40000
            }
        }
    }
}
'
```

Count distinct values of the `repositoryUrl` field for all the kustomize transformer dirs in the index:
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
        "repositoryUrl_count" : {
            "cardinality" : {
                "field" : "repositoryUrl",
                "precision_threshold": 40000
            }
        }
    }
}
'
```

List all the github repositories including kustomization files and kustomize resource files,
and how many kustomization files and kustomize resource files each github repository includes
(the github repository including the most kustomization files is listed first):
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "aggs" : {
        "repositoryUrl" : {
            "terms" : {
              "field" : "repositoryUrl",
              "size": 2082
            }
        }
    }
}
'
```

List the top 20 Github repositories including the most amount of kustomization files:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
             "filter": [
                { "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*" }}
             ]
        }
    },
    "aggs" : {
        "repositoryUrl" : {
            "terms" : {
              "field" : "repositoryUrl",
              "size": 20
            }
        }
    }
}
'
```

List the top 20 Github repositories including the most amount of kustomize resource files:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
            "must_not": {
                "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*" }
            },
            "filter": [
                { "regexp": { "fileType": "resource" }}
            ]
        }
    },
    "aggs" : {
        "repositoryUrl" : {
            "terms" : {
              "field" : "repositoryUrl",
              "size": 20
            }
        }
    }
}
'
```