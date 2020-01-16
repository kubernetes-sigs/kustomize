Count distinct values of the `repositoryUrl` field:
```
curl -X POST "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
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
curl -X POST "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
             "filter": [
                { "regexp": { "filePath": ".*/kustomization((.yaml)?|(.yml)?)/*" }}
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

Count how many Github repositories include kustomize resource files:
```
curl -X POST "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
            "must_not": {
                "regexp": { "filePath": ".*/kustomization((.yaml)?|(.yml)?)/*" }
            }
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
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
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