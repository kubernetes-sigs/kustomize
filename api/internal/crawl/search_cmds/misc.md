Check the health status of an ElasticSearch cluster:
```
curl -s -X GET "${ElasticSearchURL}:9200/_cat/health?v&pretty"
```

Check the indices in an ElasticSearch cluster:
```
curl -s "${ElasticSearchURL}:9200/_cat/indices?v"
```

Get the mapping of the index:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_mapping?pretty"
```

Delete the kustomize index from the ElasticSearch cluster (**Use this command with caution**):
```
curl -s -X DELETE "${ElasticSearchURL}:9200/${INDEXNAME}?pretty"
```

Add a new field into an existing index.
```
curl -s -X PUT "${ElasticSearchURL}:9200/${INDEXNAME}/_mapping/_doc?pretty" -H 'Content-Type: application/json' -d'
{
  "properties": {
    "fileType": {
      "type": "keyword"
    }
  }
}
'
```