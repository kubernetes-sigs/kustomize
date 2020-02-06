- [How many kustomization files on Github?](#how-many-kustomization-files-on-github)
- [How many new kustomization files were created each year on Github?](#how-many-new-kustomization-files-were-created-each-year-on-github)
- [How many kustomization files were created each month on Github?](#how-many-kustomization-files-were-created-each-month-on-github)
- [How many Github repositories include kustomization files?](#how-many-github-repositories-include-kustomization-files)
- [How many Github users include kustomization files?](#how-many-github-users-include-kustomization-files)
- [List the top 20 Github users including the most amount of kustomization files](#list-the-top-20-github-users-including-the-most-amount-of-kustomization-files)
- [How many kustomize resource files on Github?](#how-many-kustomize-resource-files-on-github)
- [How many kustomize generator files on Github?](#how-many-kustomize-generator-files-on-github)
- [How many new kustomize resource files were created each year on Github?](#how-many-new-kustomize-resource-files-were-created-each-year-on-github)
- [How many kustomize resource files were created each month on Github?](#how-many-kustomize-resource-files-were-created-each-month-on-github)
- [How many Github repositories include kustomize resource files?](#how-many-github-repositories-include-kustomize-resource-files)
- [How many Github users include kustomize resource files?](#how-many-github-users-include-kustomize-resource-files)
- [List the top 20 Github users including the most amount of kustomize resource files](#list-the-top-20-github-users-including-the-most-amount-of-kustomize-resource-files)
- [How many kustomize generator files were created each month on Github?](#how-many-kustomize-generator-files-were-created-each-month-on-github)
- [How many Github repositories including generator files?](#how-many-github-repositories-including-generator-files)
- [How many Github users including generator files?](#how-many-github-users-including-generator-files)
- [List the top 20 Github users including the most generator files](#list-the-top-20-github-users-including-the-most-generator-files)
- [How many kustomization files have a generators field?](#how-many-kustomization-files-have-a-generators-field)
- [How many kustomization roots are referred to in all the generators fields?](#how-many-kustomization-roots-are-referred-to-in-all-the-generators-fields)
- [How many Github repositories including generator directories?](#how-many-github-repositories-including-generator-directories)
- [How many Github users including generator directories?](#how-many-github-users-including-generator-directories)
- [List the top 20 Github users including the most generator directories](#list-the-top-20-github-users-including-the-most-generator-directories)
- [How many kustomize transformer files on github?](#how-many-kustomize-transformer-files-on-github)
- [How many kustomize transformer files were created each month on Github?](#how-many-kustomize-transformer-files-were-created-each-month-on-github)
- [How many Github repositories including transformer files?](#how-many-github-repositories-including-transformer-files)
- [How many Github users including transformer files?](#how-many-github-users-including-transformer-files)
- [List the top 20 Github users including the most transformer files](#list-the-top-20-github-users-including-the-most-transformer-files)
- [How many kustomization files have a transformers field?](#how-many-kustomization-files-have-a-transformers-field)
- [How many kustomization roots are referred to in all the transformers fields?](#how-many-kustomization-roots-are-referred-to-in-all-the-transformers-fields)
- [How many Github repositories including transformer directories?](#how-many-github-repositories-including-transformer-directories)
- [How many Github users including transformer directories?](#how-many-github-users-including-transformer-directories)
- [List the top 20 Github users including the most transformer directories](#list-the-top-20-github-users-including-the-most-transformer-directories)
<!-- toc -->

## How many kustomization files on Github?
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "filter": [
       { "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*" }}
      ]
    }
  }
}
'  | grep "\"hits\" : {" -A1 | grep total | awk '{print $3}' | awk -F, '{print $1}'
```



## How many new kustomization files were created each year on Github? 
Column `Year`:
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
' | grep key_as_string | awk -F\" '{print $4}' | awk -F\- '{print $1}'
```

Column `New Kustomization Files`:
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
' | grep doc_count | awk '{print $3}'
```

## How many kustomization files were created each month on Github? 
Column `Month`:
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
' | grep key_as_string | awk -F\" '{print $4}' | cut -dT -f1
```

Column `New Kustomization Files`:
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
' | grep doc_count | awk '{print $3}'
```

## How many Github repositories include kustomization files?
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
' | grep "\"value\"" | awk '{print $3}'
```

## How many Github users include kustomization files?
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
' | grep "\"value\"" | awk '{print $3}'
```

## List the top 20 Github users including the most amount of kustomization files 
Column `Github User`:
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
' | grep "\"key\"" | awk -F\" '{print $4}'
```

Column `Kustomization File Count`:
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
' | grep "\"doc_count\"" | awk '{print $3}'
```

## How many kustomize resource files on Github?
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
' | grep "\"hits\" : {" -A1 | grep total | awk '{print $3}' | awk -F, '{print $1}'
```

## How many new kustomize resource files were created each year on Github?
Column `Year`:
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
' | grep key_as_string | awk -F\" '{print $4}' | awk -F\- '{print $1}'
```

Column `New Kustomize Resource Files`:
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
' | grep "\"doc_count\"" | awk '{print $3}'
```

## How many kustomize resource files were created each month on Github? 
Column `Month`:
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
' | grep key_as_string | awk -F\" '{print $4}' | awk -FT '{print $1}'
```

Column `New Kustomize Resource Files`:
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
' | grep "\"doc_count\"" | awk '{print $3}'
```

## How many Github repositories include kustomize resource files?
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
' | grep "\"value\"" | awk '{print $3}'
```

## How many Github users include kustomize resource files?
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
' | grep "\"value\"" | awk '{print $3}'
```

## List the top 20 Github users including the most amount of kustomize resource files
Column `Github User`:
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
' | grep "\"key\"" | awk -F\" '{print $4}'
```

Column `Resource File Count`:
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
' | grep "\"doc_count\"" | awk '{print $3}'
```

## How many kustomize generator files on Github?
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
' | grep "\"hits\" : {" -A1 | grep total | awk '{print $3}' | awk -F, '{print $1}'
```

## How many kustomize generator files were created each month on Github? 
Column `Month`:
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
' | grep key_as_string | awk -F\" '{print $4}' | awk -FT '{print $1}'
```

Column `New Kustomize Transformer Files`:
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
' | grep "\"doc_count\"" | awk '{print $3}'
```

## How many Github repositories including generator files?
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
' | grep "\"value\"" | awk '{print $3}'
```

## How many Github users including generator files?
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
' | grep "\"value\"" | awk '{print $3}'
```

## List the top 20 Github users including the most generator files
Column `Github User`:
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
' | grep "\"key\"" | awk -F\" '{print $4}'
```

Column `Generator File Count`:
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
' | grep "\"doc_count\"" | awk '{print $3}'
```

## How many kustomization files have a generators field?
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
'  | grep "\"hits\" : {" -A1 | grep total | awk '{print $3}' | awk -F, '{print $1}'
```

## How many kustomization roots are referred to in all the generators fields?
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
'  | grep "\"hits\" : {" -A1 | grep total | awk '{print $3}' | awk -F, '{print $1}'
```

## How many Github repositories including generator directories?
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
' | grep "\"value\"" | awk '{print $3}'
```

## How many Github users including generator directories?
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
' | grep "\"value\"" | awk '{print $3}'
```

## List the top 20 Github users including the most generator directories
Column `Github User`:
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
' | grep "\"key\"" | awk -F\" '{print $4}'
```

Column `Generator Dir Count`:
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
' | grep "\"doc_count\"" | awk '{print $3}'
```

## How many kustomize transformer files on github?
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
' | grep "\"hits\" : {" -A1 | grep total | awk '{print $3}' | awk -F, '{print $1}'
```

## How many kustomize transformer files were created each month on Github? 
Column `Month`:
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
' | grep key_as_string | awk -F\" '{print $4}' | awk -FT '{print $1}'
```

Column `New Kustomize Transformer Files`:
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
' | grep "\"doc_count\"" | awk '{print $3}'
```

## How many Github repositories including transformer files?
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
' | grep "\"value\"" | awk '{print $3}'
```

## How many Github users including transformer files?
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
' | grep "\"value\"" | awk '{print $3}'
```

## List the top 20 Github users including the most transformer files
Column `Github User`:
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
' | grep "\"key\"" | awk -F\" '{print $4}'
```

Column `Transformer File Count`:
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
' | grep "\"doc_count\"" | awk '{print $3}'
```

## How many kustomization files have a transformers field?
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
'  | grep "\"hits\" : {" -A1 | grep total | awk '{print $3}' | awk -F, '{print $1}'
```

## How many kustomization roots are referred to in all the transformers fields?
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
'  | grep "\"hits\" : {" -A1 | grep total | awk '{print $3}' | awk -F, '{print $1}'
```

## How many Github repositories including transformer directories?
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
' | grep "\"value\"" | awk '{print $3}'
```

## How many Github users including transformer directories?
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
' | grep "\"value\"" | awk '{print $3}'
```

## List the top 20 Github users including the most transformer directories
Column `Github User`:
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
              "field" : "user",
              "size": 20
            }
        }
    }
}
' | grep "\"key\"" | awk -F\" '{print $4}'
```

Column `Transformer Dir Count`:
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
              "field" : "user",
              "size": 20
            }
        }
    }
}
' | grep "\"doc_count\"" | awk '{print $3}'
```