Find out the largest value of the `creationTime` field:
```
curl -X POST "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "aggs" : {
        "max_creationTime" : { "max" : { "field" : "creationTime" } }
    }
}
'
```

Find out the smallest value of the `creationTime` field:
```
curl -X POST "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "aggs" : {
        "min_creationTime" : { "min" : { "field" : "creationTime" } }
    }
}
'
```

Find out the smallest value of the `creationTime` field of all the kustomization files:
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
    "min_creationTime" : { "min" : { "field" : "creationTime" } }
  }
}
'
```

Find out the smallest value of the `creationTime` field of all kustomize resource files:
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
    "min_creationTime" : { "min" : { "field" : "creationTime" } }
  }
}
'
```

Query all the documents whose `creationTime` <= `2016-07-29T17:38:26.000Z`:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "range": {
      "creationTime": {
       "lte": "2016-07-29T17:38:26.000Z"
      }
    }
  }
}
'
```

Query all the documents whose `creationTime` falls within the specific range:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "range": {
      "creationTime": {
       "gte": "2016-07-29T17:38:26.000Z",
       "lte": "2016-08-29T17:38:26.000Z"
      }
    }
  }
}
'
```

Aggregate how many new kustomization files were added into Github each month:
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
        "newFiles_over_time" : {
            "date_histogram" : {
                "field" : "creationTime",
                "interval" : "month"
            }
        }
    }
}
'
```

Aggregate how many new kustomize resource files were added into Github each month:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
            "must_not": [
                { "regexp": { "filePath": ".*/kustomization((.yaml)?|(.yml)?)/*" }}
            ]
        }
    },
    "aggs" : {
        "newFiles_over_time" : {
            "date_histogram" : {
                "field" : "creationTime",
                "interval" : "month"
            }
        }
    }
}
'
```

Aggregate how many new kustomization files were added into Github each year:
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
        "newFiles_over_time" : {
            "date_histogram" : {
                "field" : "creationTime",
                "interval" : "year"
            }
        }
    }
}
'
```

Aggregate how many new kustomize resource files were added into Github each year:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
            "must_not": [
                { "regexp": { "filePath": ".*/kustomization((.yaml)?|(.yml)?)/*" }}
            ]
        }
    },
    "aggs" : {
        "newFiles_over_time" : {
            "date_histogram" : {
                "field" : "creationTime",
                "interval" : "year"
            }
        }
    }
}
'
```