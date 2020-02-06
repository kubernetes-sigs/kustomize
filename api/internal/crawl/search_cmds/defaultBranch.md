Count distinct values of the `defaultBranch` field:
```
curl -s -X POST "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "aggs" : {
        "defaultBranch_count" : {
            "cardinality" : {
                "field" : "defaultBranch",
                "precision_threshold": 40000
            }
        }
    }
}
'
```

List all the github branches where kustomization files and kustomize resource files live, 
and how many kustomization files and kustomize resource files live in each branch:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
    "aggs" : {
        "defaultBranch" : {
            "terms" : {
              "field" : "defaultBranch",
              "size": 41
            }
        }
    }
}
'
```