Find out the largest value of the `creationTime` field:
```
curl -s -X POST "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "aggs" : {
        "max_creationTime" : { "max" : { "field" : "creationTime" } }
    }
}
'
```

Find out the smallest value of the `creationTime` field:
```
curl -s -X POST "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "aggs" : {
        "min_creationTime" : { "min" : { "field" : "creationTime" } }
    }
}
'
```

Find out the smallest value of the `creationTime` field of all the kustomization files:
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
    "min_creationTime" : { "min" : { "field" : "creationTime" } }
  }
}
'
```

Find out the smallest value of the `creationTime` field of all kustomize resource files:
```
curl -s -X POST "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "must_not": {
        "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }
      }, 
      "filter": [
        { "regexp": { "fileType": "resource"  }}
      ]
    }
  },
  "aggs" : {
    "min_creationTime" : { "min" : { "field" : "creationTime" } }
  }
}
'
```

Find out the smallest value of the `creationTime` field of all kustomize generator files:
```
curl -s -X POST "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "must_not": {
        "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }
      }, 
      "filter": [
        { "regexp": { "fileType": "generator"  }}
      ]
    }
  },
  "aggs" : {
    "min_creationTime" : { "min" : { "field" : "creationTime" } }
  }
}
'
```

Find out the smallest value of the `creationTime` field of all kustomize transformer files:
```
curl -s -X POST "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "must_not": {
        "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }
      }, 
      "filter": [
        { "regexp": { "fileType": "transformer"  }}
      ]
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
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
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
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
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

Query all the kustomization files whose `creationTime` falls within the specific range:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "size": 20,
  "query": {
    "bool": {
      "filter": {
        "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*" }
      },
      "must": {
        "range": {
          "creationTime": {
            "gte": "2017-09-24T15:49:57.000Z",
            "lte": "2017-09-24T15:49:57.000Z"
          }
        }
      }
    }
  }
}
'
```

Aggregate how many new kustomization files were added into Github each month:
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
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
            "must_not": [
                { "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*" }}
            ],
            "filter": [
                { "regexp": { "fileType": "resource" }}
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

Aggregate how many new kustomize generator files were added into Github each month:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
            "must_not": [
                { "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*" }}
            ],
            "filter": [
                { "regexp": { "fileType": "generator" }}
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

Aggregate how many new kustomize transformer files were added into Github each month:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
            "must_not": [
                { "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*" }}
            ],
            "filter": [
                { "regexp": { "fileType": "transformer" }}
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
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
            "must_not": [
                { "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*" }}
            ],
            "filter": [
                { "regexp": { "fileType": "resource" }}
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

Aggregate how many new kustomize generator files were added into Github each year:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
            "must_not": [
                { "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*" }}
            ],
            "filter": [
                { "regexp": { "fileType": "generator" }}
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

Aggregate how many new kustomize transformer files were added into Github each year:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?size=0&pretty" -H 'Content-Type: application/json' -d'
{
    "query": {
        "bool": {
            "must_not": [
                { "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*" }}
            ],
            "filter": [
                { "regexp": { "fileType": "transformer" }}
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

Find the generator files created within the given time range:
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
      },
      "must": {
        "range": {
          "creationTime": {
           "gte": "2019-04-26T16:40:02.000Z",
           "lte": "2019-04-26T16:40:02.000Z"
          }
        }
      }
    }
  }
}
'
```

Find the transformer files created within the given time range:
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
      },
      "must": {
        "range": {
          "creationTime": {
           "gte": "2019-04-26T16:40:02.000Z",
           "lte": "2019-04-26T16:40:02.000Z"
          }
        }
      }
    }
  }
}
'
```