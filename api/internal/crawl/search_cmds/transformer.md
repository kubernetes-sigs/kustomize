Find all the trasnformer files whose `kinds` field includes `HelmValues`, and
only output certain fields of each document:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "size": 200,
  "_source": {
    "includes": ["kinds", "repositoryUrl", "defaultBranch", "filePath"]
  },
  "query": {
    "bool": {
      "filter": [
       { "regexp": { "fileType": "transformer" }}
      ],
      "must_not": {
        "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }
      },
      "must": {
        "match" : {
          "kinds" : {
            "query" : "HelmValues"
          }
        }
      }
    }
  }
}
'
```