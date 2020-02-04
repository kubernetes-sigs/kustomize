Find all the generator files whose `kinds` field includes `ChartRenderer`, and
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
       { "regexp": { "fileType": "generator" }}
      ],
      "must_not": {
        "regexp": { "filePath": "(.*/)?kustomization((.yaml)?|(.yml)?)(/)*"  }
      },
      "must": {
        "match" : {
          "kinds" : {
            "query" : "ChartRenderer"
          }
        }
      }
    }
  }
}
'
```