Search for all the kustomize resource files including a Deployment object:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "match" : {
      "kinds" : {
        "query" : "Deployment"
      }
    }
  }
}
'
```

Search for all the kustomize resource files including a Deployment object, but only
including the `kinds` field in the result:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "_source": {
    "includes": ["kinds"]
  },
  "query": {
    "match" : {
      "kinds" : {
        "query" : "Deployment"
      }
    }
  }
}
'
```

Search for all the kustomize resource files including both a Deployment object and 
a Service object:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "match" : {
      "kinds" : {
        "query" : "Deployment Service",
        "operator" : "and"
      }
    }
  }
}
'
```

Count the number of documents including Deployment and the number of documents 
including Service:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "size": 0,
  "aggs" : {
    "messages" : {
      "filters" : {
        "filters" : {
          "Deployment" : { "match" : { "kinds" : "Deployment"   }},
          "Service" : { "match" : { "kinds" : "Service" }}
        }
      }
    }
  }
}
'
```

Search for all the kustomization files involving CRDs:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "size": 10000,
  "query": {
    "match" : {
      "identifiers" : {
        "query" : "crds"
      }
    }
  }
}
'
```

Search for all the kustomization files defining configMapGenerator:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "size": 10000,
  "query": {
    "match" : {
      "identifiers" : {
        "query" : "configMapGenerator"
      }
    }
  }
}
'
```

Search for all the documents having a `kind` field:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "filter": [
       { "match" : { "identifiers" : { "query" : "kind" }}}
      ]
    }
  }
}
'
```

Search for all the kuostmization files having a `kind` field:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "bool": {
      "filter": [
       { "regexp": { "filePath": ".*/kustomization((.yaml)?|(.yml)?)" }},
       { "match" : { "identifiers" : { "query" : "kind" }}}
      ]
    }
  }
}
'
```

Search for all the kustomization files defining the `generatorOptions:disableNameSuffixHash` feature:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "match" : {
      "identifiers" : {
        "query" : "generatorOptions:disableNameSuffixHash"
      }
    }
  }
}
'
```